// Package agent provides functions for marshaling and unmarshaling messages.
package agent

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/dtn7/cboring"
)

var (
	// ErrUnknownType is returned by Marshal when the given value is not a
	// known message type.
	ErrUnknownType = errors.New("unknown message type")
	// ErrUnknownTag is returned by Unmarshal when the encoded type tag does
	// not correspond to a known message type.
	ErrUnknownTag = errors.New("unknown message tag")
)

// Payload type tags used to discriminate the message type on the wire.
// Commands are directed to an agent; events are emitted by an agent.
const (
	tagStartWorker uint64 = iota + 1
	tagStopWorker
	tagWorkerInput
	tagWorkerStarted
	tagWorkerStopped
	tagWorkerStatus
	tagWorkerOutput
	tagAgentInfo
)

// Marshal serializes the given message into a CBOR-encoded byte slice. It
// accepts either a value or a pointer of a known message type.
//
// The message is encoded as a CBOR array of two elements:
//
//	[type tag (uint), payload]
func Marshal(message any) ([]byte, error) {
	var tag uint64
	var payload cboring.CborMarshaler

	switch v := message.(type) {
	case StartWorker:
		tag, payload = tagStartWorker, &v
	case *StartWorker:
		tag, payload = tagStartWorker, v
	case StopWorker:
		tag, payload = tagStopWorker, &v
	case *StopWorker:
		tag, payload = tagStopWorker, v
	case WorkerInput:
		tag, payload = tagWorkerInput, &v
	case *WorkerInput:
		tag, payload = tagWorkerInput, v
	case WorkerStarted:
		tag, payload = tagWorkerStarted, &v
	case *WorkerStarted:
		tag, payload = tagWorkerStarted, v
	case WorkerStopped:
		tag, payload = tagWorkerStopped, &v
	case *WorkerStopped:
		tag, payload = tagWorkerStopped, v
	case WorkerStatus:
		tag, payload = tagWorkerStatus, &v
	case *WorkerStatus:
		tag, payload = tagWorkerStatus, v
	case WorkerOutput:
		tag, payload = tagWorkerOutput, &v
	case *WorkerOutput:
		tag, payload = tagWorkerOutput, v
	case AgentInfo:
		tag, payload = tagAgentInfo, &v
	case *AgentInfo:
		tag, payload = tagAgentInfo, v
	default:
		return nil, ErrUnknownType
	}

	var buf bytes.Buffer

	err := cboring.WriteArrayLength(2, &buf)
	if err != nil {
		return nil, err
	}

	err = cboring.WriteUInt(tag, &buf)
	if err != nil {
		return nil, err
	}

	err = payload.MarshalCbor(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes the given CBOR-encoded byte slice into a message.
// The returned value can be type-asserted to the specific message type.
func Unmarshal(b []byte) (any, error) {
	r := bytes.NewReader(b)

	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return nil, err
	}
	if l != 2 {
		return nil, errors.New("invalid message array length")
	}

	tag, err := cboring.ReadUInt(r)
	if err != nil {
		return nil, err
	}

	var payload any
	switch tag {
	case tagStartWorker:
		var m StartWorker
		err = m.UnmarshalCbor(r)
		payload = m
	case tagStopWorker:
		var m StopWorker
		err = m.UnmarshalCbor(r)
		payload = m
	case tagWorkerInput:
		var m WorkerInput
		err = m.UnmarshalCbor(r)
		payload = m
	case tagWorkerStarted:
		var m WorkerStarted
		err = m.UnmarshalCbor(r)
		payload = m
	case tagWorkerStopped:
		var m WorkerStopped
		err = m.UnmarshalCbor(r)
		payload = m
	case tagWorkerStatus:
		var m WorkerStatus
		err = m.UnmarshalCbor(r)
		payload = m
	case tagWorkerOutput:
		var m WorkerOutput
		err = m.UnmarshalCbor(r)
		payload = m
	case tagAgentInfo:
		var m AgentInfo
		err = m.UnmarshalCbor(r)
		payload = m
	default:
		return nil, ErrUnknownTag
	}
	if err != nil {
		return nil, err
	}

	if r.Len() != 0 {
		return nil, errors.New("unexpected trailing bytes")
	}

	return payload, nil
}

// CmdStartWorkerSubject returns the NATS subject for sending a start worker command to an agent.
func CmdStartWorkerSubject(gatewayID, agentID, workerID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".CMD.WORKER." + workerID + ".START"
}

// CmdStopWorkerSubject returns the NATS subject for sending a stop worker command to an agent.
func CmdStopWorkerSubject(gatewayID, agentID, workerID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".CMD.WORKER." + workerID + ".STOP"
}

// CmdWriteStdinSubject returns the NATS subject for sending stdin to a worker via an agent.
func CmdWriteStdinSubject(gatewayID, agentID, workerID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".CMD.WORKER." + workerID + ".STDIN"
}

// EvtStartWorkerSubject returns the NATS subject for a worker start event.
func EvtStartWorkerSubject(gatewayID, agentID, workerID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".EVT.WORKER." + workerID + ".START"
}

// EvtStopWorkerSubject returns the NATS subject for a worker stop event.
func EvtStopWorkerSubject(gatewayID, agentID, workerID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".EVT.WORKER." + workerID + ".STOP"
}

// EvtWorkerStatusSubject returns the NATS subject for a worker status event.
func EvtWorkerStatusSubject(gatewayID, agentID, workerID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".EVT.WORKER." + workerID + ".STATUS"
}

// EvtWorkerStdoutSubject returns the NATS subject for a worker stdout event.
func EvtWorkerStdoutSubject(gatewayID, agentID, workerID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".EVT.WORKER." + workerID + ".STDOUT"
}

// EvtAgentInfoSubject returns the NATS subject for an agent info event.
func EvtAgentInfoSubject(gatewayID, agentID string) string {
	return "FJ.GATEWAY." + gatewayID + ".AGENT." + agentID + ".EVT.INFO"
}

// Subject holds the identifiers encoded in an agent subject. WorkerID is empty
// for subjects that do not target a specific worker, such as the agent info
// event.
type Subject struct {
	GatewayID string
	AgentID   string
	WorkerID  string
}

// ParseSubject extracts the identifiers from an agent subject. It reports false
// for subjects that do not match the agent layout, so a foreign or malformed
// subject yields no identifiers rather than misattributed ones.
func ParseSubject(subject string) (Subject, bool) {
	parts := strings.Split(subject, ".")
	if len(parts) < 7 ||
		parts[0] != "FJ" ||
		parts[1] != "GATEWAY" ||
		parts[3] != "AGENT" {
		return Subject{}, false
	}

	category := parts[5]
	if category != "CMD" && category != "EVT" {
		return Subject{}, false
	}

	s := Subject{
		GatewayID: parts[2],
		AgentID:   parts[4],
	}

	if parts[6] == "WORKER" {
		if len(parts) < 9 {
			return Subject{}, false
		}
		s.WorkerID = parts[7]
	}

	return s, true
}

// writeStringSlice writes a slice of strings as a definite-length CBOR array.
func writeStringSlice(ss []string, w io.Writer) error {
	err := cboring.WriteArrayLength(uint64(len(ss)), w)
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// readStringSlice reads a definite-length CBOR array of strings. An empty
// array is decoded as a nil slice.
func readStringSlice(r io.Reader) ([]string, error) {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return nil, err
	}

	if l == 0 {
		return nil, nil
	}

	ss := make([]string, l)
	for i := range ss {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return nil, err
		}
		ss[i] = s
	}

	return ss, nil
}

// writeError writes an error as a CBOR text string. A nil error is encoded as
// an empty string.
func writeError(e error, w io.Writer) error {
	var msg string
	if e != nil {
		msg = e.Error()
	}
	return cboring.WriteTextString(msg, w)
}

// readError reads an error from a CBOR text string. An empty string is decoded
// as a nil error.
func readError(r io.Reader) (error, error) {
	msg, err := cboring.ReadTextString(r)
	if err != nil {
		return nil, err
	}

	if msg == "" {
		return nil, nil
	}

	return errors.New(msg), nil
}
