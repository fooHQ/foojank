package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetCommandResponse is a response to a GetCommandRequest.
type GetCommandResponse struct {
	Command AgentCommand
	Error   error
}

// MarshalCbor encodes the response as a CBOR array: [command, error].
func (m *GetCommandResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = m.Command.MarshalCbor(w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *GetCommandResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	var command AgentCommand
	err = command.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.Command = command

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
