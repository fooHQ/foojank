package router

import (
	"context"
	"github.com/foojank/foojank/internal/services/client/reader"
	"github.com/nats-io/nats.go"
	"github.com/urfave/cli/v2"
	"strings"
)

type Arguments struct {
	App        *cli.App
	Connection *nats.Conn
	InputCh    <-chan reader.Message
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
			data := msg.Data()
			args := strings.Split(data, " ")
			err := s.args.App.Run(args)
			if err != nil {
				continue
			}
			_ = msg.Reply()

		case <-ctx.Done():
			return nil
		}
	}
}
