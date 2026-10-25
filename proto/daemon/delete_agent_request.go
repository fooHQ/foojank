package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// DeleteAgentRequest is a request to delete an agent.
// AgentRef identifies the agent.
type DeleteAgentRequest struct {
	AgentRef string
}

// MarshalCbor encodes the request as a CBOR array: [agentRef].
func (m *DeleteAgentRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.AgentRef, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *DeleteAgentRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	agentRef, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.AgentRef = agentRef

	return nil
}
