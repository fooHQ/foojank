package proto

import (
	capnplib "capnproto.org/go/capnp/v3"

	"github.com/foohq/foojank/proto/capnp"
)

func Marshal(data any) ([]byte, error) {
	switch v := data.(type) {
	case StartWorkerRequest:
		return marshalStartWorkerRequest(v)
	case StartWorkerResponse:
		return marshalStartWorkerResponse(v)
	case StopWorkerRequest:
		return marshalStopWorkerRequest(v)
	case StopWorkerResponse:
		return marshalStopWorkerResponse(v)
	case UpdateWorkerStatus:
		return marshalUpdateWorkerStatus(v)
	case UpdateWorkerStdio:
		return marshalUpdateWorkerStdio(v)
	case UpdateClientInfo:
		return marshalUpdateClientInfo(v)
	}
	return nil, ErrUnknownMessage
}

func marshalStartWorkerRequest(data StartWorkerRequest) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewStartWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = m.SetCommand(data.Command)
	if err != nil {
		return nil, err
	}

	argsList, err := newTextList(msg.Segment(), data.Args)
	if err != nil {
		return nil, err
	}

	err = m.SetArgs(argsList)
	if err != nil {
		return nil, err
	}

	envList, err := newTextList(msg.Segment(), data.Env)
	if err != nil {
		return nil, err
	}

	err = m.SetEnv(envList)
	if err != nil {
		return nil, err
	}

	err = msg.Content().SetStartWorkerRequest(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
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

func marshalUpdateClientInfo(data UpdateClientInfo) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewUpdateClientInfo(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = m.SetUsername(data.Username)
	if err != nil {
		return nil, err
	}

	err = m.SetHostname(data.Hostname)
	if err != nil {
		return nil, err
	}

	err = m.SetSystem(data.System)
	if err != nil {
		return nil, err
	}

	err = m.SetAddress(data.Address)
	if err != nil {
		return nil, err
	}

	err = msg.Content().SetUpdateClientInfo(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func newMessage() (capnp.Message, error) {
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

func newTextList(segment *capnplib.Segment, ss []string) (capnplib.TextList, error) {
	tl, err := capnplib.NewTextList(segment, int32(len(ss)))
	if err != nil {
		return capnplib.TextList{}, err
	}

	for i, s := range ss {
		err := tl.Set(i, s)
		if err != nil {
			return capnplib.TextList{}, err
		}
	}

	return tl, nil
}
