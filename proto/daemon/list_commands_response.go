package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListCommandsResponse is a response to a ListCommandsRequest.
// Commands is keyed by command name.
type ListCommandsResponse struct {
	Commands map[string]AgentCommand
	Error    error
}

// MarshalCbor encodes the response as a CBOR array: [commands, error].
func (m *ListCommandsResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = writeSortedMap(m.Commands, (*AgentCommand).MarshalCbor, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *ListCommandsResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	commands, err := readSortedMap((*AgentCommand).UnmarshalCbor, r)
	if err != nil {
		return err
	}
	m.Commands = commands

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
