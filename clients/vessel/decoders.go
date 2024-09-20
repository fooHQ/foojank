package vessel

import (
	"capnproto.org/go/capnp/v3"
	"github.com/foojank/foojank/proto"
)

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
