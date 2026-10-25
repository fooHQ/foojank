package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetRoleRequest is a request to retrieve a role.
// RoleRef identifies the role.
type GetRoleRequest struct {
	RoleRef string
}

// MarshalCbor encodes the request as a CBOR array: [roleRef].
func (m *GetRoleRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.RoleRef, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *GetRoleRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	roleRef, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.RoleRef = roleRef

	return nil
}
