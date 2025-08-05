package decoder

import (
	"context"

	"github.com/foohq/foojank/internal/vessel/consumer"
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
			parsed, err := proto.ParseAction(msg.Data())
			if err != nil {
				log.Debug("cannot decode message", "error", err)
				continue
			}

			var data any
			switch v := parsed.(type) {
			case proto.CreateJobRequest:
				data = v
			case proto.CancelJobRequest:
				data = v
			default:
				log.Debug("cannot decode message: unknown message", "message", parsed)
				continue
			}

			select {
			case s.args.OutputCh <- Message{
				msg:  msg,
				data: data,
			}:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
