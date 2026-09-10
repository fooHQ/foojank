package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListAgentsResponse is a response to a ListAgentsRequest.
type ListAgentsResponse struct {
	Agents []Agent
	Error  error
}

// MarshalCbor encodes the response as a CBOR array: [agents, error].
func (m *ListAgentsResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = writeAgentSlice(m.Agents, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *ListAgentsResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	agents, err := readAgentSlice(r)
	if err != nil {
		return err
	}
	m.Agents = agents

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
