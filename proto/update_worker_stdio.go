package proto

import "github.com/foohq/foojank/proto/capnp"

// UpdateWorkerStdio contains data from a worker's stdout or stdin.
type UpdateWorkerStdio struct {
	Data []byte
}

func marshalUpdateWorkerStdio(data UpdateWorkerStdio) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewUpdateWorkerStdio(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = m.SetData(data.Data)
	if err != nil {
		return nil, err
	}

	err = msg.Content().SetUpdateWorkerStdio(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func unmarshalUpdateWorkerStdio(message capnp.Message) (UpdateWorkerStdio, error) {
	v, err := message.Content().UpdateWorkerStdio()
	if err != nil {
		return UpdateWorkerStdio{}, err
	}

	data, err := v.Data()
	if err != nil {
		return UpdateWorkerStdio{}, err
	}

	return UpdateWorkerStdio{
		Data: data,
	}, nil
}
