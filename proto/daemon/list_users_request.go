package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListUsersRequest is a request to list users.
type ListUsersRequest struct{}

// MarshalCbor encodes the request as an empty CBOR array.
func (m *ListUsersRequest) MarshalCbor(w io.Writer) error {
	return cboring.WriteArrayLength(0, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *ListUsersRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 0 {
		return errors.New("invalid message array length")
	}

	return nil
}
