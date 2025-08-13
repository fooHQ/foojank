package publisher

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/vessel/log"
)

type Arguments struct {
	Connection jetstream.JetStream
	InputCh    <-chan Message
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
	for {
		select {
		case msg := <-s.args.InputCh:
			_, err := s.args.Connection.Publish(ctx, msg.subject, msg.data)
			if err != nil {
				log.Debug("cannot publish to stdout subject", "error", err)
				continue
			}

		case <-ctx.Done():
			return nil
		}
	}
}
