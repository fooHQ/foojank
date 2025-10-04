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
	defer log.Debug("Service stopped", "service", "vessel", "id", s.args.ID)

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
	defer func() {
		_ = natsFS.Close()
	}()

	filesystems := map[string]risoros.FS{
		"file": fileFS,
		"mem":  memFS,
		"nats": natsFS,
	}

	consumerOutCh := make(chan message.Msg)
	publisherInCh := make(chan message.Msg)
	publisherOutCh := make(chan message.Msg)
	termCh := make(chan struct{})

	var wg sync.WaitGroup

	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := consumer(consumerCtx, s.args.Consumer, consumerOutCh)
		if err != nil {
			log.Debug("Consumer error", "error", err)
		}
		termCh <- struct{}{}
	}()

	workManagerCtx, workManagerCancel := context.WithCancel(context.Background())
	defer workManagerCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := workmanager.New(workmanager.Arguments{
			ID:          s.args.ID,
			Templates:   s.args.Templates,
			Filesystems: filesystems,
			InputCh:     consumerOutCh,
			OutputCh:    publisherInCh,
		}).Start(workManagerCtx)
		if err != nil {
			log.Debug("WorkManager error", "error", err)
		}
		termCh <- struct{}{}
	}()

	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	defer monitorCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := monitor(monitorCtx, s.args.Connection, s.args.Templates.Render(subjects.ConnInfo, s.args.ID), publisherInCh)
		if err != nil {
			log.Debug("Monitor error", "error", err)
		}
		termCh <- struct{}{}
	}()

	publisherCtx, publisherCancel := context.WithCancel(context.Background())
	defer publisherCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := publisher(publisherCtx, s.args.Connection, publisherInCh, publisherOutCh)
		if err != nil {
			log.Debug("Publisher error", "error", err)
		}
		termCh <- struct{}{}
	}()

	ackerCtx, ackerCancel := context.WithCancel(context.Background())
	defer ackerCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := acker(ackerCtx, publisherOutCh)
		if err != nil {
			log.Debug("Acker error", "error", err)
		}
		termCh <- struct{}{}
	}()

	cancels := []context.CancelFunc{
		consumerCancel,
		workManagerCancel,
		monitorCancel,
		publisherCancel,
		ackerCancel,
	}

	select {
	case <-ctx.Done():
		for _, cancel := range cancels {
			cancel()
			<-termCh
		}
	case <-termCh:
		// If an error occurs in one of the services, cancel all services without waiting for them to finish.
		// Some messages may be lost in the process.
		for _, cancel := range cancels {
			cancel()
		}
	}

	wg.Wait()

	return nil
}

var _ message.Msg = consumerMessage{}

type consumerMessage struct {
	msg  jetstream.Msg
	data any
}

func (m consumerMessage) ID() string {
	return m.msg.Headers().Get(nats.MsgIdHdr)
}

func (m consumerMessage) Subject() string {
	return m.msg.Subject()
}

func (m consumerMessage) Data() any {
	return m.data
}

func (m consumerMessage) Ack() error {
	return m.msg.Ack()
}

func consumer(ctx context.Context, consumer jetstream.Consumer, outputCh chan message.Msg) error {
	log.Debug("Service started", "service", "vessel.consumer")
	defer log.Debug("Service stopped", "service", "vessel.consumer")

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

			data, err := proto.Unmarshal(msg.Data())
			if err != nil {
				log.Debug("Cannot decode a message", "error", err)
				_ = msg.Ack()
				continue
			}

			err = forwardMessage(outputCh, consumerMessage{
				msg:  msg,
				data: data,
			})
			if err != nil {
				log.Debug("Cannot forward a message", "error", err)
				continue
			}
		}
	}()

	<-ctx.Done()
	msgs.Stop()
	wg.Wait()

	return nil
}

// TODO: messages should be aggregated so that CreateWorkerRequest can be canceled by StopWorkerRequest.

type publisherMessage struct {
	msg  message.Msg
	data jetstream.PubAckFuture
}

func (m publisherMessage) ID() string {
	return m.msg.ID()
}

func (m publisherMessage) Subject() string {
	return m.msg.Subject()
}

func (m publisherMessage) Data() any {
	return m.data
}

func (m publisherMessage) Ack() error {
	return m.msg.Ack()
}

func publisher(ctx context.Context, conn jetstream.JetStream, inputCh <-chan message.Msg, outputCh chan<- message.Msg) error {
	log.Debug("Service started", "service", "vessel.publisher")
	defer log.Debug("Service stopped", "service", "vessel.publisher")

	opts := []jetstream.PublishOpt{
		jetstream.WithRetryAttempts(3),
		jetstream.WithRetryWait(250 * time.Millisecond),
	}

loop:
	for {
		select {
		case msg := <-inputCh:
			data, err := proto.Marshal(msg.Data())
			if err != nil {
				log.Debug("Cannot encode a message", "error", err)
				_ = msg.Ack()
				continue
			}

			pubAck, err := conn.PublishAsync(
				msg.Subject(),
				data,
				opts...,
			)
			if err != nil {
				log.Debug("Cannot publish a message", "subject", msg.Subject(), "error", err)
				continue
			}

			err = forwardMessage(outputCh, publisherMessage{
				msg:  msg,
				data: pubAck,
			})
			if err != nil {
				log.Debug("Cannot forward a message", "error", err)
				continue
			}

		case <-ctx.Done():
			break loop
		}
	}

	select {
	case <-conn.PublishAsyncComplete():
	case <-time.After(5 * time.Second):
		log.Debug("Some messages were not published", "error", "timeout while waiting for publish to complete")
	}

	return nil
}

func acker(ctx context.Context, inputCh <-chan message.Msg) error {
	log.Debug("Service started", "service", "vessel.acker")
	defer log.Debug("Service stopped", "service", "vessel.acker")

	for {
		select {
		case msg := <-inputCh:
			ack, ok := msg.Data().(jetstream.PubAckFuture)
			if !ok {
				log.Debug("Cannot cast to PubAckFuture")
				return errors.New("invalid data")
			}

			// Waiting for ctx is not necessary as the publisher always sets a timeout on how long a single request can take.
			select {
			case <-ack.Ok():
				_ = msg.Ack()
			case err := <-ack.Err():
				log.Debug("Cannot publish a message", "subject", msg.Subject(), "error", err)
			}

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
			case <-time.After(3 * time.Second):
				log.Debug("Timeout while waiting to write to output channel")
				continue
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

func forwardMessage(outputCh chan<- message.Msg, msg message.Msg) error {
	select {
	case outputCh <- msg:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("timeout")
	}
}
