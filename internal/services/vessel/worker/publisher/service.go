package publisher

import (
	"context"
	"github.com/foojank/foojank/internal/log"
	"github.com/nats-io/nats.go"
)

type Arguments struct {
	Connection *nats.Conn
	Subject    string
	InputCh    <-chan []byte
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
		case b := <-s.args.InputCh:
			msg := &nats.Msg{
				Subject: s.args.Subject,
				Data:    b,
			}
			err := s.args.Connection.PublishMsg(msg)
			if err != nil {
				log.Debug("cannot publish to stdout subject: %v", err)
				continue
			}

		case <-ctx.Done():
			return nil
		}
	}
}
