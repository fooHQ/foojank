package vessel

import (
	"capnproto.org/go/capnp/v3"
	"github.com/foojank/foojank/proto"
)

func NewCreateWorkerRequest() ([]byte, error) {
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, err
	}

	msg, err := proto.NewRootMessage(seg)
	if err != nil {
		return nil, err
	}

	msgCreateWorker, err := proto.NewCreateWorkerRequest(seg)
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
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, err
	}

	msg, err := proto.NewRootMessage(seg)
	if err != nil {
		return nil, err
	}

	msgDestroyWorker, err := proto.NewDestroyWorkerRequest(seg)
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
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, err
	}

	msg, err := proto.NewRootMessage(seg)
	if err != nil {
		return nil, err
	}

	msgGetWorker, err := proto.NewGetWorkerRequest(seg)
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
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, err
	}

	msg, err := proto.NewRootMessage(seg)
	if err != nil {
		return nil, err
	}

	msgExecute, err := proto.NewExecuteRequest(seg)
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
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, err
	}

	msg, err := proto.NewRootDummyRequest(seg)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}
