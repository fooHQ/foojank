package proto

import (
	"errors"

	capnplib "capnproto.org/go/capnp/v3"

	"github.com/foohq/foojank/proto/capnp"
)

var (
	ErrUnknownMessage = errors.New("unknown message")
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

func Unmarshal(b []byte) (any, error) {
	message, err := parseMessage(b)
	if err != nil {
		return nil, err
	}

	content := message.Content()
	switch {
	case content.HasStartWorkerRequest():
		return unmarshalStartWorkerRequest(message)

	case content.HasStartWorkerResponse():
		return unmarshalStartWorkerResponse(message)

	case content.HasStopWorkerRequest():
		return unmarshalStopWorkerRequest(message)

	case content.HasStopWorkerResponse():
		return unmarshalStopWorkerResponse(message)

	case content.HasUpdateWorkerStatus():
		return unmarshalUpdateWorkerStatus(message)

	case content.HasUpdateWorkerStdio():
		return unmarshalUpdateWorkerStdio(message)

	case content.HasUpdateClientInfo():
		return unmarshalUpdateClientInfo(message)
	}

	return nil, ErrUnknownMessage
}

func StartWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.StartWorkerT, agentID, workerID)
}

func StopWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.StopWorkerT, agentID, workerID)
}

func WriteWorkerStdinSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.WriteWorkerStdinT, agentID, workerID)
}

func WriteWorkerStdoutSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.WriteWorkerStdoutT, agentID, workerID)
}

func UpdateWorkerStatusSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.UpdateWorkerStatusT, agentID, workerID)
}

func UpdateClientInfoSubject(agentID string) string {
	return replaceStringPlaceholders(capnp.UpdateClientInfoT, agentID)
}

func ReplyMessageSubject(agentID, msgID string) string {
	return replaceStringPlaceholders(capnp.ReplyMessageT, agentID, msgID)
}

func replaceStringPlaceholders(s string, values ...string) string {
	result := s
	valIndex := 0

	for valIndex < len(values) {
		found := false
		for i := 0; i < len(result)-1; i++ {
			if result[i] == '%' && result[i+1] == 's' {
				result = result[:i] + values[valIndex] + result[i+2:]
				valIndex++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return result
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

func parseMessage(b []byte) (capnp.Message, error) {
	capMsg, err := capnplib.Unmarshal(b)
	if err != nil {
		return capnp.Message{}, err
	}

	message, err := capnp.ReadRootMessage(capMsg)
	if err != nil {
		return capnp.Message{}, err
	}

	return message, nil
}

func textListToStringSlice(list capnplib.TextList) ([]string, error) {
	if list.Len() == 0 {
		return nil, nil
	}
	result := make([]string, 0, list.Len())
	for i := 0; i < list.Len(); i++ {
		v, err := list.At(i)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
