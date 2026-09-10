package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListUsersResponse is a response to a ListUsersRequest.
type ListUsersResponse struct {
	Users []User
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [users, error].
func (m *ListUsersResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = writeUserSlice(m.Users, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *ListUsersResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	users, err := readUserSlice(r)
	if err != nil {
		return err
	}
	m.Users = users

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
