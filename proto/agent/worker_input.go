package agent

import (
	"io"

	"github.com/dtn7/cboring"
)

// WorkerInput is a command carrying data for a worker's stdin.
type WorkerInput struct {
	Data []byte
}

// MarshalCbor encodes the data as a CBOR byte string.
func (m *WorkerInput) MarshalCbor(w io.Writer) error {
	return cboring.WriteByteString(m.Data, w)
}

// UnmarshalCbor decodes the data from its CBOR representation.
func (m *WorkerInput) UnmarshalCbor(r io.Reader) error {
	data, err := cboring.ReadByteString(r)
	if err != nil {
		return err
	}

	m.Data = data

	return nil
}
