package agent

import (
	"io"

	"github.com/dtn7/cboring"
)

// WorkerOutput is an event carrying data from a worker's stdout.
type WorkerOutput struct {
	Data []byte
}

// MarshalCbor encodes the data as a CBOR byte string.
func (m *WorkerOutput) MarshalCbor(w io.Writer) error {
	return cboring.WriteByteString(m.Data, w)
}

// UnmarshalCbor decodes the data from its CBOR representation.
func (m *WorkerOutput) UnmarshalCbor(r io.Reader) error {
	data, err := cboring.ReadByteString(r)
	if err != nil {
		return err
	}

	m.Data = data

	return nil
}
