package decoder

import (
	"context"

	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/connector"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	InputCh     <-chan connector.Message
	DataSubject string
	DataCh      chan<- Message
	StdinCh     chan<- Message
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
			if msg.Subject() == s.args.DataSubject {
				parsed, err := proto.ParseAction(msg.Data())
				if err != nil {
					log.Debug("cannot decode worker action message", "error", err)
					_ = msg.ReplyError(errcodes.ErrInvalidMessage, "", nil)
					continue
				}

				var data any
				switch v := parsed.(type) {
				case proto.ExecuteRequest:
					data = ExecuteRequest{
						Args:     v.Args,
						FilePath: v.FilePath,
					}

				default:
					log.Debug("invalid scheduler action message", "message", parsed)
					_ = msg.ReplyError(errcodes.ErrInvalidAction, "", nil)
					continue
				}

				select {
				case s.args.DataCh <- Message{
					ctx:        ctx,
					req:        msg,
					responseCh: responseCh,
					data:       data,
				}:
				case <-ctx.Done():
					return nil
				}
			} else {
				data := msg.Data()

				select {
				case s.args.StdinCh <- Message{
					ctx:  ctx,
					req:  msg,
					data: data,
				}:
				case <-ctx.Done():
					return nil
				}
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
			case ExecuteResponse:
				b, err = proto.NewExecuteResponse(v.Code)

			default:
				log.Debug("invalid worker response message", "message", msg.Data())
				_ = msg.Request().ReplyError(errcodes.ErrInvalidResponse, "", nil)
				continue
			}
			if err != nil {
				log.Debug("cannot create a worker response message", "error", err)
				_ = msg.Request().ReplyError(errcodes.ErrNewResponseFailed, err.Error(), nil)
				continue
			}

			_ = msg.Request().Reply(b)

		case <-ctx.Done():
			return nil
		}
	}
}
