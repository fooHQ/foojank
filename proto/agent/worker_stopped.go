package agent

import (
	"io"
)

// WorkerStopped is an event reporting the outcome of a StopWorker command.
type WorkerStopped struct {
	Error error
}

// MarshalCbor encodes the event, storing the error as a text string.
func (m *WorkerStopped) MarshalCbor(w io.Writer) error {
	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the event from its CBOR representation.
func (m *WorkerStopped) UnmarshalCbor(r io.Reader) error {
	respErr, err := readError(r)
	if err != nil {
		return err
	}

	m.Error = respErr

	return nil
}
