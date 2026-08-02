// Package gateway provides functions for marshaling and unmarshaling messages.
package gateway

import (
	"bytes"
	"errors"
	"io"
	"sort"

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
const (
	tagRegisterAgentRequest uint64 = iota + 1
	tagRegisterAgentResponse
	tagUnregisterAgentRequest
	tagUnregisterAgentResponse
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
	case RegisterAgentRequest:
		tag, payload = tagRegisterAgentRequest, &v
	case *RegisterAgentRequest:
		tag, payload = tagRegisterAgentRequest, v
	case RegisterAgentResponse:
		tag, payload = tagRegisterAgentResponse, &v
	case *RegisterAgentResponse:
		tag, payload = tagRegisterAgentResponse, v
	case UnregisterAgentRequest:
		tag, payload = tagUnregisterAgentRequest, &v
	case *UnregisterAgentRequest:
		tag, payload = tagUnregisterAgentRequest, v
	case UnregisterAgentResponse:
		tag, payload = tagUnregisterAgentResponse, &v
	case *UnregisterAgentResponse:
		tag, payload = tagUnregisterAgentResponse, v
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
	case tagRegisterAgentRequest:
		var m RegisterAgentRequest
		err = m.UnmarshalCbor(r)
		payload = m
	case tagRegisterAgentResponse:
		var m RegisterAgentResponse
		err = m.UnmarshalCbor(r)
		payload = m
	case tagUnregisterAgentRequest:
		var m UnregisterAgentRequest
		err = m.UnmarshalCbor(r)
		payload = m
	case tagUnregisterAgentResponse:
		var m UnregisterAgentResponse
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

// RegisterAgentSubject returns the NATS subject for registering an agent with the gateway.
func RegisterAgentSubject(gatewayID string) string {
	return "FJ.GATEWAY." + gatewayID + ".RPC.AGENT.REGISTER"
}

// UnregisterAgentSubject returns the NATS subject for unregistering an agent from the gateway.
func UnregisterAgentSubject(gatewayID string) string {
	return "FJ.GATEWAY." + gatewayID + ".RPC.AGENT.UNREGISTER"
}

// writeStringMap writes a map of strings as a definite-length CBOR map. Keys
// are written in sorted order for deterministic output.
func writeStringMap(m map[string]string, w io.Writer) error {
	err := cboring.WriteMapPairLength(uint64(len(m)), w)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := cboring.WriteTextString(k, w)
		if err != nil {
			return err
		}
		err = cboring.WriteTextString(m[k], w)
		if err != nil {
			return err
		}
	}

	return nil
}

// readStringMap reads a definite-length CBOR map of strings. An empty map is
// decoded as a nil map.
func readStringMap(r io.Reader) (map[string]string, error) {
	l, err := cboring.ReadMapPairLength(r)
	if err != nil {
		return nil, err
	}

	if l == 0 {
		return nil, nil
	}

	m := make(map[string]string, l)
	for range l {
		k, err := cboring.ReadTextString(r)
		if err != nil {
			return nil, err
		}

		v, err := cboring.ReadTextString(r)
		if err != nil {
			return nil, err
		}

		m[k] = v
	}

	return m, nil
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
