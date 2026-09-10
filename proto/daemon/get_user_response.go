package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetUserResponse is a response to a GetUserRequest.
type GetUserResponse struct {
	User  User
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [user, error].
func (m *GetUserResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = m.User.MarshalCbor(w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *GetUserResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	var user User
	err = user.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.User = user

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
