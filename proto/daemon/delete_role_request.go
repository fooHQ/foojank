package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// DeleteRoleRequest is a request to delete a role.
type DeleteRoleRequest struct {
	Name string
}

// MarshalCbor encodes the request as a CBOR array: [name].
func (m *DeleteRoleRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.Name, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *DeleteRoleRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	name, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.Name = name

	return nil
}
