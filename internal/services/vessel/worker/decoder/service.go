package decoder

import (
	"capnproto.org/go/capnp/v3"
	"context"
	"github.com/foojank/foojank/internal/log"
	"github.com/foojank/foojank/internal/services/vessel/worker/connector"
	"github.com/foojank/foojank/proto"
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
	responseCh := make(chan MessageResponse, 65535)

	for {
		select {
		case msg := <-s.args.InputCh:
			if msg.Subject() == s.args.DataSubject {
				data, err := s.decodeData(msg)
				if err != nil {
					_ = msg.ReplyError("400", err.Error(), nil)
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
			// TODO: refactor response encoding!!!
			msgErr := msg.Error()
			if msgErr != nil {
				_ = msg.Request().ReplyError(msgErr.Code, msgErr.Description, nil)
				continue
			}

			rootMsg, err := newRootMessage()
			if err != nil {
				log.Debug("cannot create a root message: %v", err)
				_ = msg.Request().ReplyError("500", err.Error(), nil)
				continue
			}

			data := msg.Data()
			switch v := data.(type) {
			case ExecuteResponse:
				rootMsg, err = newExecuteResponse(rootMsg, v.Code)
				if err != nil {
					log.Debug("cannot create a response message: %v", err)
					_ = msg.Request().ReplyError("500", err.Error(), nil)
					continue
				}

			default:
				log.Debug("unknown response type: %T", v)
				continue
			}

			b, err := rootMsg.Message().Marshal()
			if err != nil {
				log.Debug("cannot marshal the message: %v", err)
				_ = msg.Request().ReplyError("500", err.Error(), nil)
				continue
			}

			_ = msg.Request().Reply(b)

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) decodeData(msg connector.Message) (any, error) {
	capMsg, err := capnp.Unmarshal(msg.Data())
	if err != nil {
		log.Debug("cannot decode input data: %v", err)
		return nil, err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		log.Debug("cannot decode scheduler action message: %v", err)
		return nil, err
	}

	action := message.Action()
	var data any
	switch {
	case action.HasExecute():
		v, _ := action.Execute()
		// TODO: validate error?
		b, _ := v.Data()
		data = ExecuteRequest{
			Data: b,
		}
	}

	return data, nil
}

func newRootMessage() (proto.Message, error) {
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return proto.Message{}, err
	}

	rootMsg, err := proto.NewRootMessage(seg)
	if err != nil {
		return proto.Message{}, err
	}

	return rootMsg, nil
}

func newExecuteResponse(root proto.Message, code int64) (proto.Message, error) {
	respMsg, err := proto.NewExecuteResponse(root.Segment())
	if err != nil {
		return proto.Message{}, err
	}

	respMsg.SetCode(code)

	err = root.Response().SetExecute(respMsg)
	if err != nil {
		return proto.Message{}, err
	}

	return root, nil
}
