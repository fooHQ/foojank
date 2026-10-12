package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// JobOutputResponse is a response to a JobOutputRequest.
// Data is the next chunk of output after the requested offset. Empty Data
// with a nil Error means no further output is available yet.
type JobOutputResponse struct {
	Data  []byte
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [data, error].
func (m *JobOutputResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteByteString(m.Data, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *JobOutputResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	data, err := cboring.ReadByteString(r)
	if err != nil {
		return err
	}
	m.Data = data

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
