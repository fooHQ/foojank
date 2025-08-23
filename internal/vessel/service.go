package vessel

import (
	"context"
	"errors"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/message"
	"github.com/foohq/foojank/internal/vessel/workmanager"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	Connection jetstream.JetStream
	Stream     string // TODO: convert to jetstream.Stream
	Consumer   string
	Subject    string
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
	log.Debug("Service started", "service", "vessel")
	defer log.Debug("Service stopped", "service", "vessel")

	consumerOutCh := make(chan message.Msg)
	decoderOutCh := make(chan message.Msg)
	encoderInCh := make(chan any)
	publisherInCh := make(chan []byte)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return consumer(groupCtx, s.args.Connection, s.args.Stream, s.args.Consumer, encoderInCh, consumerOutCh)
	})

	group.Go(func() error {
		return decoder(groupCtx, consumerOutCh, decoderOutCh)
	})

	group.Go(func() error {
		return workmanager.New(workmanager.Arguments{
			Connection: s.args.Connection,
			Stream:     s.args.Stream,
			InputCh:    decoderOutCh,
		}).Start(groupCtx)
	})

	group.Go(func() error {
		return encoder(groupCtx, encoderInCh, publisherInCh)
	})

	group.Go(func() error {
		return publisher(groupCtx, s.args.Connection, s.args.Subject, publisherInCh)
	})

	return group.Wait()
}

type consumerMessage struct {
	msg     jetstream.Msg
	replyCh chan<- any
}

func (m consumerMessage) Data() any {
	return m.msg.Data()
}

func (m consumerMessage) Ack() error {
	return m.msg.Ack()
}

func (m consumerMessage) Reply(ctx context.Context, data any) error {
	if m.replyCh == nil {
		return errors.New("message does not expect a reply")
	}
	select {
	case m.replyCh <- data:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func consumer(ctx context.Context, conn jetstream.JetStream, stream, consumer string, replyCh chan any, outputCh chan message.Msg) error {
	log.Debug("Service started", "service", "vessel.consumer")
	defer log.Debug("Service stopped", "service", "vessel.consumer")

	c, err := conn.Consumer(ctx, stream, consumer)
	if err != nil {
		log.Debug("Cannot initialize durable consumer", "name", consumer, "stream", stream, "error", err)
		return err
	}

	msgs, err := c.Messages()
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
				msg:     msg,
				replyCh: replyCh,
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

type decoderMessage struct {
	msg  message.Msg
	data any
}

func (m decoderMessage) Ack() error {
	return m.msg.Ack()
}

func (m decoderMessage) Reply(ctx context.Context, data any) error {
	return m.msg.Reply(ctx, data)
}

func (m decoderMessage) Data() any {
	return m.data
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

func encoder(ctx context.Context, inputCh <-chan any, outputCh chan<- []byte) error {
	log.Debug("Service started", "service", "vessel.encoder")
	defer log.Debug("Service stopped", "service", "vessel.encoder")

	for {
		select {
		case msg := <-inputCh:
			b, err := proto.Marshal(msg)
			if err != nil {
				log.Debug("Cannot encode a message", "error", err)
				continue
			}

			select {
			case outputCh <- b:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func publisher(ctx context.Context, conn jetstream.JetStream, subject string, inputCh <-chan []byte) error {
	log.Debug("Service started", "service", "vessel.publisher")
	defer log.Debug("Service stopped", "service", "vessel.publisher")

	for {
		select {
		case msg := <-inputCh:
			_, err := conn.Publish(ctx, subject, msg)
			if err != nil {
				log.Debug("Cannot publish a message", "error", err)
				continue
			}

		case <-ctx.Done():
			return nil
		}
	}
}
