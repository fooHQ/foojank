package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListRolesResponse is a response to a ListRolesRequest.
type ListRolesResponse struct {
	Roles []Role
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [roles, error].
func (m *ListRolesResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = writeRoleSlice(m.Roles, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *ListRolesResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	roles, err := readRoleSlice(r)
	if err != nil {
		return err
	}
	m.Roles = roles

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
