package consumer

import (
	"context"
	"errors"
	"sync"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/vessel/log"
)

type Arguments struct {
	Connection     jetstream.JetStream
	Stream         string
	Consumer       string
	Durable        bool
	FilterSubjects []string
	ReplyCh        chan<- any
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
	var consumer jetstream.Consumer
	var err error
	if s.args.Durable {
		consumer, err = s.args.Connection.Consumer(ctx, s.args.Stream, s.args.Consumer)
	} else {
		consumer, err = s.args.Connection.CreateConsumer(ctx, s.args.Stream, jetstream.ConsumerConfig{
			Name:           s.args.Consumer,
			DeliverPolicy:  jetstream.DeliverAllPolicy,
			AckPolicy:      jetstream.AckExplicitPolicy,
			MaxAckPending:  1,
			FilterSubjects: s.args.FilterSubjects,
		})
	}
	if err != nil {
		log.Debug("cannot initialize consumer", "error", err)
		return err
	}

	msgs, err := consumer.Messages()
	if err != nil {
		log.Debug("cannot obtain message context", "error", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		forwardMessages(ctx, msgs, s.args.ReplyCh, s.args.OutputCh)
	}()

	<-ctx.Done()
	msgs.Stop()
	wg.Wait()

	return nil
}

func forwardMessages(ctx context.Context, msgs jetstream.MessagesContext, replyCh chan<- any, outputCh chan<- Message) {
	for {
		msg, err := msgs.Next()
		if err != nil {
			if errors.Is(err, jetstream.ErrMsgIteratorClosed) {
				return
			}
			continue
		}

		outMsg := Message{
			msg:     msg,
			replyCh: replyCh,
		}
		select {
		case outputCh <- outMsg:
		case <-ctx.Done():
			return
		}
	}
}
