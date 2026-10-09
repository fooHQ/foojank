package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListRolesRequest is a request to list roles.
type ListRolesRequest struct{}

// MarshalCbor encodes the request as an empty CBOR array.
func (m *ListRolesRequest) MarshalCbor(w io.Writer) error {
	return cboring.WriteArrayLength(0, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *ListRolesRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 0 {
		return errors.New("invalid message array length")
	}

	return nil
}
