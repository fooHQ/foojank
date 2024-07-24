package runner

import (
	"context"
	"github.com/foojank/foojank/internal/services/connector"
	"github.com/risor-io/risor"
)

type Arguments struct {
	InputCh <-chan connector.Message
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
			src := string(msg.Data)
			// TODO: configure risor
			// TODO: return stdout/stderr
			_, err := risor.Eval(ctx, src)
			if err != nil {
				// TODO: use ReplyError
				msg.Reply(ctx, []byte(err.Error()))
				continue
			}

			msg.Reply(ctx, []byte("OK"))

		case <-ctx.Done():
			return nil
		}
	}
}
