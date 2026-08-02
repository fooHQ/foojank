package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// UnregisterAgentRequest is a request to unregister an agent from the gateway.
type UnregisterAgentRequest struct {
	AgentID string
}

// MarshalCbor encodes the request as a CBOR array: [agentID].
func (m *UnregisterAgentRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.AgentID, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *UnregisterAgentRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	agentID, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.AgentID = agentID

	return nil
}
