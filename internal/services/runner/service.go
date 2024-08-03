package runner

import (
	"context"
	"github.com/foojank/foojank/internal/services/connector"
	"github.com/risor-io/risor"
	risoros "github.com/risor-io/risor/os"
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
			stdout := risoros.NewBufferFile(nil)
			ros := NewOS(ctx, stdout)
			ctx = risoros.WithOS(ctx, ros)
			src := string(msg.Data())
			_, err := risor.Eval(ctx, src)
			if err != nil {
				_ = msg.ReplyError("400", err.Error(), nil)
				continue
			}

			_ = msg.Reply(stdout.Bytes())

		case <-ctx.Done():
			return nil
		}
	}
}
