package proto

import "github.com/foohq/foojank/proto/capnp"

// UpdateWorkerStatus is used to update the status of a worker.
type UpdateWorkerStatus struct {
	Status int64
}

func marshalUpdateWorkerStatus(data UpdateWorkerStatus) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewUpdateWorkerStatus(msg.Segment())
	if err != nil {
		return nil, err
	}

	m.SetStatus(data.Status)

	err = msg.Content().SetUpdateWorkerStatus(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func unmarshalUpdateWorkerStatus(message capnp.Message) (UpdateWorkerStatus, error) {
	v, err := message.Content().UpdateWorkerStatus()
	if err != nil {
		return UpdateWorkerStatus{}, err
	}

	return UpdateWorkerStatus{
		Status: v.Status(),
	}, nil
}
