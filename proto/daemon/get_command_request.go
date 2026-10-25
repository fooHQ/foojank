package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetCommandRequest is a request to retrieve one command of an agent.
// AgentRef identifies the agent. Name is the command name.
type GetCommandRequest struct {
	AgentRef string
	Name     string
}

// MarshalCbor encodes the request as a CBOR array: [agentRef, name].
func (m *GetCommandRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.AgentRef, m.Name} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *GetCommandRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.AgentRef, &m.Name} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	return nil
}
