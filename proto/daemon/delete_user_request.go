package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// DeleteUserRequest is a request to delete a user.
// UserRef identifies the user.
type DeleteUserRequest struct {
	UserRef string
}

// MarshalCbor encodes the request as a CBOR array: [userRef].
func (m *DeleteUserRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.UserRef, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *DeleteUserRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	userRef, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.UserRef = userRef

	return nil
}
