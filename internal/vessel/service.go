package vessel

import (
	"context"
	"errors"
	"os"
	"os/user"
	"runtime"
	"sync"
	"time"

	memfs "github.com/foohq/ren-memfs"
	natsfs "github.com/foohq/ren-natsfs"
	localfs "github.com/foohq/ren/filesystems/local"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/message"
	"github.com/foohq/foojank/internal/vessel/subjects"
	"github.com/foohq/foojank/internal/vessel/workmanager"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	ID          string
	Connection  jetstream.JetStream
	Stream      jetstream.Stream
	Consumer    jetstream.Consumer
	ObjectStore jetstream.ObjectStore
	Templates   subjects.Templates
}

type Service struct {
	args Arguments
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	log.Debug("Service started", "service", "vessel", "id", s.args.ID)
	defer log.Debug("Service stopped", "service", "vessel")

	fileFS, err := localfs.NewFS()
	if err != nil {
		log.Debug("Cannot instantiate local fs", "error", err)
		return err
	}

	memFS, err := memfs.NewFS()
	if err != nil {
		log.Debug("Cannot instantiate mem fs", "error", err)
		return err
	}

	natsFS, err := natsfs.NewFS(ctx, s.args.ObjectStore)
	if err != nil {
		log.Debug("Cannot instantiate nats fs", "error", err)
		return err
	}
	defer natsFS.Close()

	filesystems := map[string]risoros.FS{
		"file": fileFS,
		"mem":  memFS,
		"nats": natsFS,
	}

	consumerOutCh := make(chan message.Msg)
	decoderOutCh := make(chan message.Msg)
	encoderInCh := make(chan message.Msg)
	publisherInCh := make(chan message.Msg)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return consumer(groupCtx, s.args.Consumer, consumerOutCh)
	})

	group.Go(func() error {
		return decoder(groupCtx, consumerOutCh, decoderOutCh)
	})

	group.Go(func() error {
		return workmanager.New(workmanager.Arguments{
			ID:          s.args.ID,
			Templates:   s.args.Templates,
			Filesystems: filesystems,
			InputCh:     decoderOutCh,
			OutputCh:    encoderInCh,
		}).Start(groupCtx)
	})

	group.Go(func() error {
		return monitor(groupCtx, s.args.Connection, s.args.Templates.Render(subjects.ConnInfo, s.args.ID), encoderInCh)
	})

	group.Go(func() error {
		return encoder(groupCtx, encoderInCh, publisherInCh)
	})

	group.Go(func() error {
		return publisher(groupCtx, s.args.Connection, publisherInCh)
	})

	return group.Wait()
}

var _ message.Msg = consumerMessage{}

type consumerMessage struct {
	msg jetstream.Msg
}

func (m consumerMessage) ID() string {
	return m.msg.Headers().Get(nats.MsgIdHdr)
}

func (m consumerMessage) Subject() string {
	return m.msg.Subject()
}

func (m consumerMessage) Data() any {
	return m.msg.Data()
}

func (m consumerMessage) Ack() error {
	return m.msg.Ack()
}

func consumer(ctx context.Context, consumer jetstream.Consumer, outputCh chan message.Msg) error {
	log.Debug("Service started", "service", "vessel.consumer")
	defer log.Debug("Service stopped", "service", "vessel.consumer")

	// TODO: consumer should be able to go offline and online to avoid creating long standing connections.

	msgs, err := consumer.Messages()
	if err != nil {
		log.Debug("Cannot obtain message context", "error", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			msg, err := msgs.Next()
			if err != nil {
				if errors.Is(err, jetstream.ErrMsgIteratorClosed) {
					return
				}
				continue
			}

			select {
			case outputCh <- consumerMessage{
				msg: msg,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	msgs.Stop()
	wg.Wait()

	return nil
}

// TODO: messages should be aggregated so that CreateWorkerRequest can be canceled by StopWorkerRequest.

var _ message.Msg = decoderMessage{}

type decoderMessage struct {
	msg  message.Msg
	data any
}

func (m decoderMessage) ID() string {
	return m.msg.ID()
}

func (m decoderMessage) Subject() string {
	return m.msg.Subject()
}

func (m decoderMessage) Data() any {
	return m.data
}

func (m decoderMessage) Ack() error {
	return m.msg.Ack()
}

func decoder(ctx context.Context, inputCh <-chan message.Msg, outputCh chan<- message.Msg) error {
	log.Debug("Service started", "service", "vessel.decoder")
	defer log.Debug("Service stopped", "service", "vessel.decoder")

	for {
		select {
		case msg := <-inputCh:
			v, ok := msg.Data().([]byte)
			if !ok {
				log.Debug("Cannot decode a message", "error", errors.New("cannot cast to []byte"))
				_ = msg.Ack()
				continue
			}

			decoded, err := proto.Unmarshal(v)
			if err != nil {
				log.Debug("Cannot decode a message", "error", err)
				_ = msg.Ack()
				continue
			}

			select {
			case outputCh <- decoderMessage{
				msg:  msg,
				data: decoded,
			}:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

var _ message.Msg = encoderMessage{}

type encoderMessage struct {
	msg  message.Msg
	data any
}

func (m encoderMessage) ID() string {
	return m.msg.ID()
}

func (m encoderMessage) Subject() string {
	return m.msg.Subject()
}

func (m encoderMessage) Data() any {
	return m.data
}

func (m encoderMessage) Ack() error {
	return m.msg.Ack()
}

func encoder(ctx context.Context, inputCh <-chan message.Msg, outputCh chan<- message.Msg) error {
	log.Debug("Service started", "service", "vessel.encoder")
	defer log.Debug("Service stopped", "service", "vessel.encoder")

	for {
		select {
		case msg := <-inputCh:
			data, err := proto.Marshal(msg.Data())
			if err != nil {
				log.Debug("Cannot encode a message", "error", err)
				_ = msg.Ack()
				continue
			}

			select {
			case outputCh <- encoderMessage{
				msg:  msg,
				data: data,
			}:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func publisher(ctx context.Context, conn jetstream.JetStream, inputCh <-chan message.Msg) error {
	log.Debug("Service started", "service", "vessel.publisher")
	defer log.Debug("Service stopped", "service", "vessel.publisher")

	for {
		select {
		case msg := <-inputCh:
			data, ok := msg.Data().([]byte)
			if !ok {
				log.Debug("Cannot convert data to byte slice")
				_ = msg.Ack()
				continue
			}

			_, err := conn.Publish(ctx, msg.Subject(), data)
			if err != nil {
				log.Debug("Cannot publish a message", "subject", msg.Subject(), "error", err)
				continue
			}

			_ = msg.Ack()

		case <-ctx.Done():
			return nil
		}
	}
}

var _ message.Msg = monitorMessage{}

type monitorMessage struct {
	subject string
	data    any
}

func (m monitorMessage) ID() string {
	return ""
}

func (m monitorMessage) Ack() error {
	return message.ErrUnsupported
}

func (m monitorMessage) Subject() string {
	return m.subject
}

func (m monitorMessage) Data() any {
	return m.data
}

func monitor(ctx context.Context, conn jetstream.JetStream, subject string, outputCh chan<- message.Msg) error {
	log.Debug("Service started", "service", "vessel.monitor")
	defer log.Debug("Service stopped", "service", "vessel.monitor")

	triggerCh := make(chan struct{}, 2)

	if conn.Conn().Status() == nats.CONNECTED {
		triggerCh <- struct{}{}
	}

	for {
		select {
		case status := <-conn.Conn().StatusChanged():
			log.Debug("Service status", "status", status.String())

			if status != nats.CONNECTED {
				continue
			}

			triggerCh <- struct{}{}

		// TODO: this should be configurable!
		case <-time.After(55 * time.Minute):
			if conn.Conn().Status() != nats.CONNECTED {
				continue
			}

			triggerCh <- struct{}{}

		case <-triggerCh:
			select {
			case outputCh <- monitorMessage{
				subject: subject,
				data: proto.UpdateClientInfo{
					Username: getUsername(),
					Hostname: getHostname(),
					System:   getSystem(),
					Address:  getAddress(conn.Conn()),
				},
			}:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func getUsername() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	return usr.Username
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func getSystem() string {
	return runtime.GOOS
}

func getAddress(nc *nats.Conn) string {
	ip, err := nc.GetClientIP()
	if err != nil {
		return ""
	}
	return ip.String()
}
