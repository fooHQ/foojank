package proto

import (
	"errors"

	"github.com/foohq/foojank/proto/capnp"
)

type StartWorkerResponse struct {
	Error error
}

func marshalStartWorkerResponse(data StartWorkerResponse) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewStartWorkerResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	if data.Error != nil {
		err = m.SetError(data.Error.Error())
		if err != nil {
			return nil, err
		}
	}

	err = msg.Content().SetStartWorkerResponse(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func unmarshalStartWorkerResponse(message capnp.Message) (StartWorkerResponse, error) {
	v, err := message.Content().StartWorkerResponse()
	if err != nil {
		return StartWorkerResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return StartWorkerResponse{}, err
	}

	var respErr error
	if errMsg != "" {
		respErr = errors.New(errMsg)
	}

	return StartWorkerResponse{
		Error: respErr,
	}, nil
}
