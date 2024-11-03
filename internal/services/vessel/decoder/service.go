package decoder

import (
	"context"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/services/vessel/connector"
	"github.com/foohq/foojank/internal/services/vessel/errcodes"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	InputCh  <-chan connector.Message
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
	responseCh := make(chan MessageResponse)

	for {
		select {
		case msg := <-s.args.InputCh:
			parsed, err := proto.ParseAction(msg.Data())
			if err != nil {
				log.Debug("cannot decode scheduler action message: %v", err)
				_ = msg.ReplyError(errcodes.ErrInvalidMessage, "", nil)
				continue
			}

			var data any
			switch v := parsed.(type) {
			case proto.CreateWorkerRequest:
				data = CreateWorkerRequest{}

			case proto.DestroyWorkerRequest:
				data = DestroyWorkerRequest{
					ID: v.ID,
				}

			case proto.GetWorkerRequest:
				data = GetWorkerRequest{
					ID: v.ID,
				}

			default:
				log.Debug("invalid scheduler action message: %T", parsed)
				_ = msg.ReplyError(errcodes.ErrInvalidAction, "", nil)
				continue
			}

			select {
			case s.args.OutputCh <- Message{
				ctx:        ctx,
				req:        msg,
				responseCh: responseCh,
				data:       data,
			}:
			case <-ctx.Done():
				return nil
			}

		case msg := <-responseCh:
			msgErr := msg.Error()
			if msgErr != nil {
				_ = msg.Request().ReplyError(msgErr.Code, msgErr.Description, nil)
				continue
			}

			var b []byte
			var err error
			switch v := msg.Data().(type) {
			case CreateWorkerResponse:
				b, err = proto.NewCreateWorkerResponse(v.ID)

			case DestroyWorkerResponse:
				b, err = proto.NewDestroyWorkerResponse()

			case GetWorkerResponse:
				b, err = proto.NewGetWorkerResponse(v.ServiceName, v.ServiceID)

			default:
				log.Debug("invalid scheduler response message: %T", msg.Data())
				_ = msg.Request().ReplyError(errcodes.ErrInvalidResponse, "", nil)
				continue
			}
			if err != nil {
				log.Debug("cannot create a scheduler response message: %v", err)
				_ = msg.Request().ReplyError(errcodes.ErrNewResponseFailed, err.Error(), nil)
				continue
			}

			_ = msg.Request().Reply(b)

		case <-ctx.Done():
			return nil
		}
	}
}
