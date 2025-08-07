package connector

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/connector/consumer"
	"github.com/foohq/foojank/internal/vessel/connector/decoder"
	"github.com/foohq/foojank/internal/vessel/connector/encoder"
	"github.com/foohq/foojank/internal/vessel/connector/publisher"
)

type Arguments struct {
	Connection jetstream.JetStream
	Stream     string
	Consumer   string
	Subject    string
	OutputCh   chan<- Message
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
	consumerOutCh := make(chan consumer.Message)
	encoderInCh := make(chan any)
	encoderOutCh := make(chan []byte)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return consumer.New(consumer.Arguments{
			Connection: s.args.Connection,
			Stream:     s.args.Stream,
			Consumer:   s.args.Consumer,
			ReplyCh:    encoderInCh,
			OutputCh:   consumerOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return decoder.New(decoder.Arguments{
			InputCh:  consumerOutCh,
			OutputCh: s.args.OutputCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return encoder.New(encoder.Arguments{
			InputCh:  encoderInCh,
			OutputCh: encoderOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return publisher.New(publisher.Arguments{
			Connection: s.args.Connection,
			Subject:    s.args.Subject,
			InputCh:    encoderOutCh,
		}).Start(groupCtx)
	})

	return group.Wait()
}
