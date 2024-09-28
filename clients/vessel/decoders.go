package vessel

import (
	"capnproto.org/go/capnp/v3"
	"github.com/foojank/foojank/proto"
)

// TODO: parse root message (exported function!) DRY!

func ParseCreateWorkerRequest(b []byte) error {
	capMsg, err := capnp.Unmarshal(b)
	if err != nil {
		return err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		return err
	}

	_, err = message.Action().CreateWorker()
	if err != nil {
		return err
	}

	return nil
}

func ParseCreateWorkerResponse(b []byte) (uint64, error) {
	capMsg, err := capnp.Unmarshal(b)
	if err != nil {
		return 0, err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		return 0, err
	}

	v, err := message.Response().CreateWorker()
	if err != nil {
		return 0, err
	}

	return v.Id(), nil
}

func ParseGetWorkerRequest(b []byte) (uint64, error) {
	capMsg, err := capnp.Unmarshal(b)
	if err != nil {
		return 0, err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		return 0, err
	}

	v, err := message.Action().GetWorker()
	if err != nil {
		return 0, err
	}

	return v.Id(), nil
}

func ParseGetWorkerResponse(b []byte) (string, string, error) {
	capMsg, err := capnp.Unmarshal(b)
	if err != nil {
		return "", "", err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		return "", "", err
	}

	v, err := message.Response().GetWorker()
	if err != nil {
		return "", "", err
	}

	serviceName, err := v.ServiceName()
	if err != nil {
		return "", "", err
	}

	serviceID, err := v.ServiceId()
	if err != nil {
		return "", "", err
	}

	return serviceName, serviceID, nil
}

func ParseExecuteRequest(b []byte) ([]byte, error) {
	capMsg, err := capnp.Unmarshal(b)
	if err != nil {
		return nil, err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		return nil, err
	}

	v, err := message.Action().Execute()
	if err != nil {
		return nil, err
	}

	data, err := v.Data()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ParseExecuteResponse(b []byte) (int64, error) {
	capMsg, err := capnp.Unmarshal(b)
	if err != nil {
		return 0, err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		return 0, err
	}

	v, err := message.Response().Execute()
	if err != nil {
		return 0, err
	}

	return v.Code(), nil
}
