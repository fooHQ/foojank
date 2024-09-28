package vessel

import (
	"capnproto.org/go/capnp/v3"
	"github.com/foojank/foojank/proto"
)

func ParseMessage(b []byte) (proto.Message, error) {
	capMsg, err := capnp.Unmarshal(b)
	if err != nil {
		return proto.Message{}, err
	}

	message, err := proto.ReadRootMessage(capMsg)
	if err != nil {
		return proto.Message{}, err
	}

	return message, nil
}

func ParseCreateWorkerRequest(b []byte) error {
	message, err := ParseMessage(b)
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
	message, err := ParseMessage(b)
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
	message, err := ParseMessage(b)
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
	message, err := ParseMessage(b)
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
	message, err := ParseMessage(b)
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
	message, err := ParseMessage(b)
	if err != nil {
		return 0, err
	}

	v, err := message.Response().Execute()
	if err != nil {
		return 0, err
	}

	return v.Code(), nil
}
