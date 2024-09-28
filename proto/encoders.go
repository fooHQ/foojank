package proto

import (
	capnplib "capnproto.org/go/capnp/v3"
	"github.com/foojank/foojank/proto/capnp"
)

func NewMessage() (capnp.Message, error) {
	arena := capnplib.SingleSegment(nil)
	_, seg, err := capnplib.NewMessage(arena)
	if err != nil {
		return capnp.Message{}, err
	}

	msg, err := capnp.NewRootMessage(seg)
	if err != nil {
		return capnp.Message{}, err
	}

	return msg, nil
}

func NewCreateWorkerRequest() ([]byte, error) {
	msg, err := NewMessage()
	if err != nil {
		return nil, err
	}

	msgCreateWorker, err := capnp.NewCreateWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msg.Action().SetCreateWorker(msgCreateWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewDestroyWorkerRequest(id uint64) ([]byte, error) {
	msg, err := NewMessage()
	if err != nil {
		return nil, err
	}

	msgDestroyWorker, err := capnp.NewDestroyWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	msgDestroyWorker.SetId(id)

	err = msg.Action().SetDestroyWorker(msgDestroyWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewGetWorkerRequest(id uint64) ([]byte, error) {
	msg, err := NewMessage()
	if err != nil {
		return nil, err
	}

	msgGetWorker, err := capnp.NewGetWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	msgGetWorker.SetId(id)

	err = msg.Action().SetGetWorker(msgGetWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewExecuteRequest(data []byte) ([]byte, error) {
	msg, err := NewMessage()
	if err != nil {
		return nil, err
	}

	msgExecute, err := capnp.NewExecuteRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	_ = msgExecute.SetData(data)

	err = msg.Action().SetExecute(msgExecute)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewDummyRequest() ([]byte, error) {
	msg, err := NewMessage()
	if err != nil {
		return nil, err
	}

	msgDummy, err := capnp.NewDummyRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msg.Action().SetDummyRequest(msgDummy)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}
