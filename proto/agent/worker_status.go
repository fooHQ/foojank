package agent

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// Worker exit status codes.
const (
	ExitSuccess         int64 = 0
	ExitFailure         int64 = 1
	ExitCommandNotFound int64 = 127
	ExitInterrupted     int64 = 130
)

// WorkerStatus is an event reporting the status of a worker.
type WorkerStatus struct {
	Status int64
	Error  error
}

// MarshalCbor encodes the event as a CBOR array: [status, error].
func (m *WorkerStatus) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteInt(m.Status, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the event from its CBOR representation.
func (m *WorkerStatus) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	status, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}

	respErr, err := readError(r)
	if err != nil {
		return err
	}

	m.Status = status
	m.Error = respErr

	return nil
}
