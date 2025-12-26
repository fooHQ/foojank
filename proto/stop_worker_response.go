package proto

import (
	"errors"

	"github.com/foohq/foojank/proto/capnp"
)

// StopWorkerResponse is a response to a StopWorkerRequest.
type StopWorkerResponse struct {
	Error error
}

func marshalStopWorkerResponse(data StopWorkerResponse) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewStopWorkerResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	if data.Error != nil {
		err = m.SetError(data.Error.Error())
		if err != nil {
			return nil, err
		}
	}

	err = msg.Content().SetStopWorkerResponse(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func unmarshalStopWorkerResponse(message capnp.Message) (StopWorkerResponse, error) {
	v, err := message.Content().StopWorkerResponse()
	if err != nil {
		return StopWorkerResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return StopWorkerResponse{}, err
	}

	var respErr error
	if errMsg != "" {
		respErr = errors.New(errMsg)
	}

	return StopWorkerResponse{
		Error: respErr,
	}, nil
}
