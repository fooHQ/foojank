package encoder

import (
	"context"
	"errors"

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
			var b []byte
			var err error
			switch v := msg.(type) {
			case proto.CreateJobResponse:
				b, err = proto.NewCreateJobResponse(v.JobID, v.StdinSubject, v.StdoutSubject, v.Error)
			case proto.CancelJobResponse:
				b, err = proto.NewCancelJobResponse(v.Error)
			case proto.UpdateJob:
				b, err = proto.NewUpdateJob(v.JobID, v.ExitStatus)
			default:
				err = errors.New("unknown message")
			}
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
