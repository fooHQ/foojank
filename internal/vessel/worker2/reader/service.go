package reader

import (
	"context"
	"errors"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	Connection jetstream.JetStream
	Stream     string
	Subject    string
	File       risoros.File
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
	decoderOutCh := make(chan []byte)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return consumer(groupCtx, s.args.Connection, s.args.Stream, s.args.Subject, consumerOutCh)
	})

	group.Go(func() error {
		return decoder(groupCtx, consumerOutCh, decoderOutCh)
	})

	group.Go(func() error {
		return fileWriter(groupCtx, decoderOutCh, s.args.File)
	})

	<-groupCtx.Done()
	_ = s.args.File.Close()

	err := group.Wait()
	if err != nil {
		log.Debug("reader stopped", "error", err)
		return err
	}

	return nil
}

type consumerMessage struct {
	msg jetstream.Msg
}

func (m consumerMessage) Data() []byte {
	return m.msg.Data()
}

func (m consumerMessage) Ack() error {
	return m.msg.Ack()
}

func consumer(ctx context.Context, conn jetstream.JetStream, stream, subject string, outputCh chan consumerMessage) error {
	c, err := conn.CreateConsumer(ctx, stream, jetstream.ConsumerConfig{
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
		FilterSubject: subject,
	})
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

func decoder(ctx context.Context, inputCh <-chan consumerMessage, outputCh chan<- []byte) error {
	for {
		select {
		case msg := <-inputCh:
			err := msg.Ack()
			if err != nil {
				log.Debug("cannot ack message", "error", err)
				continue
			}

			decoded, err := proto.Unmarshal(msg.Data())
			if err != nil {
				log.Debug("cannot decode message", "error", err)
				continue
			}

			data, ok := decoded.(proto.UpdateStdioLine)
			if !ok {
				continue
			}

			select {
			case outputCh <- []byte(data.Text):
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func fileWriter(ctx context.Context, inputCh <-chan []byte, stdin risoros.File) error {
	for {
		select {
		case msg := <-inputCh:
			_, err := stdin.Write(msg)
			if err != nil {
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
