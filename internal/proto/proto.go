// Package proto provides functions for marshaling and unmarshaling messages
// and generating NATS subjects for communication.
package proto

import (
	"errors"

	capnplib "capnproto.org/go/capnp/v3"

	capnp "github.com/foohq/foojank-proto/go/agent"
)

var (
	ErrUnknownMessage = errors.New("unknown message")
)

// Marshal serializes the given data into a byte slice.
// It supports various request and response types defined in the proto package.
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

// Unmarshal deserializes the given byte slice into a message object.
// It returns an interface{} which can be type-asserted to the specific message type.
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

// CmdStartWorkerSubject returns the NATS subject for sending a start worker command to an agent.
func CmdStartWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.CmdStartWorkerT, agentID, workerID)
}

// CmdStopWorkerSubject returns the NATS subject for sending a stop worker command to an agent.
func CmdStopWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.CmdStopWorkerT, agentID, workerID)
}

// CmdWriteStdinSubject returns the NATS subject for sending stdin to a worker via an agent.
func CmdWriteStdinSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.CmdWriteStdinT, agentID, workerID)
}

// EvtStartWorkerSubject returns the NATS subject for a worker start event.
func EvtStartWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.EvtStartWorkerT, agentID, workerID)
}

// EvtStopWorkerSubject returns the NATS subject for a worker stop event.
func EvtStopWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.EvtStopWorkerT, agentID, workerID)
}

// EvtWorkerStatusSubject returns the NATS subject for a worker status event.
func EvtWorkerStatusSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.EvtWorkerStatusT, agentID, workerID)
}

// EvtWorkerStdoutSubject returns the NATS subject for a worker stdout event.
func EvtWorkerStdoutSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.EvtWorkerStdoutT, agentID, workerID)
}

// EvtAgentInfoSubject returns the NATS subject for an agent info event.
func EvtAgentInfoSubject(agentID string) string {
	return replaceStringPlaceholders(capnp.EvtAgentInfoT, agentID)
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
