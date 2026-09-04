package agent

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// StopWorker is a command to stop a running worker.
type StopWorker struct{}

// MarshalCbor encodes the command as an empty CBOR array.
func (m *StopWorker) MarshalCbor(w io.Writer) error {
	return cboring.WriteArrayLength(0, w)
}

// UnmarshalCbor decodes the command from its CBOR representation.
func (m *StopWorker) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 0 {
		return errors.New("invalid message array length")
	}

	return nil
}
