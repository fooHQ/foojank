package proto

import "github.com/foohq/foojank/proto/capnp"

type StopWorkerRequest struct{}

func marshalStopWorkerRequest(_ StopWorkerRequest) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewStopWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msg.Content().SetStopWorkerRequest(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func unmarshalStopWorkerRequest(message capnp.Message) (StopWorkerRequest, error) {
	_, err := message.Content().StopWorkerRequest()
	if err != nil {
		return StopWorkerRequest{}, err
	}

	return StopWorkerRequest{}, nil
}
