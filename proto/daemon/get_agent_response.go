package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetAgentResponse is a response to a GetAgentRequest.
type GetAgentResponse struct {
	Agent Agent
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [agent, error].
func (m *GetAgentResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = m.Agent.MarshalCbor(w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *GetAgentResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	var agent Agent
	err = agent.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.Agent = agent

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
