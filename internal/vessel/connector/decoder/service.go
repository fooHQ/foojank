package decoder

import (
	"context"

	"github.com/foohq/foojank/internal/vessel/connector/consumer"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	InputCh  <-chan consumer.Message
	OutputCh chan<- Message
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
			decoded, err := proto.Unmarshal(msg.Data())
			if err != nil {
				log.Debug("cannot decode message", "error", err)
				_ = msg.Ack()
				continue
			}

			select {
			case s.args.OutputCh <- Message{
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
