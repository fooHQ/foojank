package decoder

import (
	"capnproto.org/go/capnp/v3"
	"context"
	"github.com/foojank/foojank/internal/log"
	"github.com/foojank/foojank/internal/services/vessel/connector"
	"github.com/foojank/foojank/proto"
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
	responseCh := make(chan MessageResponse, 65535)

	for {
		select {
		case msg := <-s.args.InputCh:
			capMsg, err := capnp.Unmarshal(msg.Data())
			if err != nil {
				log.Debug("cannot decode input data: %v", err)
				_ = msg.ReplyError("400", err.Error(), nil)
				continue
			}

			message, err := proto.ReadRootMessage(capMsg)
			if err != nil {
				log.Debug("cannot decode scheduler action message: %v", err)
				_ = msg.ReplyError("400", err.Error(), nil)
				continue
			}

			action := message.Action()
			var data any
			switch {
			case action.HasCreateWorker():
				data = CreateWorkerRequest{}
			case action.HasDestroyWorker():
				v, _ := action.DestroyWorker()
				data = DestroyWorkerRequest{
					ID: v.Id(),
				}

			case action.HasGetWorker():
				v, _ := action.GetWorker()
				data = GetWorkerRequest{
					ID: v.Id(),
				}
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

			rootMsg, err := newRootMessage()
			if err != nil {
				log.Debug("cannot create a root message: %v", err)
				_ = msg.Request().ReplyError("500", err.Error(), nil)
				continue
			}

			data := msg.Data()
			switch v := data.(type) {
			case CreateWorkerResponse:
				rootMsg, err = newCreateWorkerResponse(rootMsg, v.ID)
				if err != nil {
					log.Debug("cannot create a response message: %v", err)
					_ = msg.Request().ReplyError("500", err.Error(), nil)
					continue
				}

			case DestroyWorkerResponse:
				rootMsg, err = newDestroyWorkerResponse(rootMsg)
				if err != nil {
					log.Debug("cannot create a response message: %v", err)
					_ = msg.Request().ReplyError("500", err.Error(), nil)
					continue
				}

			case GetWorkerResponse:
				rootMsg, err = newGetWorkerResponse(rootMsg, v.ServiceID)
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

func newCreateWorkerResponse(root proto.Message, id uint64) (proto.Message, error) {
	respMsg, err := proto.NewCreateWorkerResponse(root.Segment())
	if err != nil {
		return proto.Message{}, err
	}

	respMsg.SetId(id)

	err = root.Response().SetCreateWorker(respMsg)
	if err != nil {
		return proto.Message{}, err
	}

	return root, nil
}

func newDestroyWorkerResponse(root proto.Message) (proto.Message, error) {
	respMsg, err := proto.NewDestroyWorkerResponse(root.Segment())
	if err != nil {
		return proto.Message{}, err
	}

	err = root.Response().SetDestroyWorker(respMsg)
	if err != nil {
		return proto.Message{}, err
	}

	return root, nil
}

func newGetWorkerResponse(root proto.Message, serviceID string) (proto.Message, error) {
	respMsg, err := proto.NewGetWorkerResponse(root.Segment())
	if err != nil {
		return proto.Message{}, err
	}

	err = respMsg.SetServiceId(serviceID)
	if err != nil {
		return proto.Message{}, err
	}

	err = root.Response().SetGetWorker(respMsg)
	if err != nil {
		return proto.Message{}, err
	}

	return root, nil
}
