package agent

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// StartWorker is a command to start a new worker.
type StartWorker struct {
	Command string
	Args    []string
	Env     []string
}

// MarshalCbor encodes the command as a CBOR array: [command, args, env].
func (m *StartWorker) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.Command, w)
	if err != nil {
		return err
	}

	err = writeStringSlice(m.Args, w)
	if err != nil {
		return err
	}

	return writeStringSlice(m.Env, w)
}

// UnmarshalCbor decodes the command from its CBOR representation.
func (m *StartWorker) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
		return errors.New("invalid message array length")
	}

	command, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}

	args, err := readStringSlice(r)
	if err != nil {
		return err
	}

	env, err := readStringSlice(r)
	if err != nil {
		return err
	}

	m.Command = command
	m.Args = args
	m.Env = env

	return nil
}
