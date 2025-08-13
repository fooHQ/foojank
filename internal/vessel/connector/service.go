package connector

import (
	"context"
	"errors"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	Connection     jetstream.JetStream
	Stream         string
	Consumer       string
	Durable        bool
	FilterSubjects []string
	ReplySubject   string
	OutputCh       chan<- Message
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
	consumerOutCh := make(chan consumerMessage)
	encoderInCh := make(chan any)
	publisherInCh := make(chan []byte)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return consumer(groupCtx, s.args.Connection, s.args.Stream, s.args.Consumer, encoderInCh, consumerOutCh)
	})

	group.Go(func() error {
		return decoder(groupCtx, consumerOutCh, s.args.OutputCh)
	})

	group.Go(func() error {
		return encoder(groupCtx, encoderInCh, publisherInCh)
	})

	group.Go(func() error {
		return publisher(groupCtx, s.args.Connection, s.args.ReplySubject, publisherInCh)
	})

	return group.Wait()
}

type consumerMessage struct {
	msg     jetstream.Msg
	replyCh chan<- any
}

func (m consumerMessage) Data() []byte {
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

func consumer(ctx context.Context, conn jetstream.JetStream, stream, consumer string, replyCh chan any, outputCh chan consumerMessage) error {
	c, err := conn.Consumer(ctx, stream, consumer)
	if err != nil {
		log.Debug("cannot initialize durable consumer", "name", consumer, "stream", stream, "error", err)
		return err
	}

	msgs, err := c.Messages()
	if err != nil {
		log.Debug("cannot obtain message context", "error", err)
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

func decoder(ctx context.Context, inputCh <-chan consumerMessage, outputCh chan<- Message) error {
	for {
		select {
		case msg := <-inputCh:
			decoded, err := proto.Unmarshal(msg.Data())
			if err != nil {
				log.Debug("cannot decode message", "error", err)
				_ = msg.Ack()
				continue
			}

			select {
			case outputCh <- Message{
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
	for {
		select {
		case msg := <-inputCh:
			b, err := proto.Marshal(msg)
			if err != nil {
				log.Debug("cannot encode message", "error", err)
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
	for {
		select {
		case msg := <-inputCh:
			_, err := conn.Publish(ctx, subject, msg)
			if err != nil {
				log.Debug("cannot publish message", "error", err)
				continue
			}

		case <-ctx.Done():
			return nil
		}
	}
}
