package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// JobInputResponse is a response to a JobInputRequest.
type JobInputResponse struct {
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [error].
func (m *JobInputResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *JobInputResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
