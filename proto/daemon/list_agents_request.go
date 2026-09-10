package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListAgentsRequest is a request to list agents.
type ListAgentsRequest struct{}

// MarshalCbor encodes the request as an empty CBOR array.
func (m *ListAgentsRequest) MarshalCbor(w io.Writer) error {
	return cboring.WriteArrayLength(0, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *ListAgentsRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 0 {
		return errors.New("invalid message array length")
	}

	return nil
}
