package encoder

import (
	"context"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	InputCh  <-chan any
	OutputCh chan<- []byte
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
			b, err := proto.Marshal(msg)
			if err != nil {
				log.Debug("cannot encode message", "error", err)
				continue
			}

			select {
			case s.args.OutputCh <- b:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
