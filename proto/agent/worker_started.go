package agent

import (
	"io"
)

// WorkerStarted is an event reporting the outcome of a StartWorker command.
type WorkerStarted struct {
	Error error
}

// MarshalCbor encodes the event, storing the error as a text string.
func (m *WorkerStarted) MarshalCbor(w io.Writer) error {
	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the event from its CBOR representation.
func (m *WorkerStarted) UnmarshalCbor(r io.Reader) error {
	respErr, err := readError(r)
	if err != nil {
		return err
	}

	m.Error = respErr

	return nil
}
