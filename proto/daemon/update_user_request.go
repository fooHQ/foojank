package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// UpdateUserRequest edits the user named Name.
//
// Description replaces the description when IsDescription is true. SetPrivileges
// are added to the user's privilege set and UnsetPrivileges are removed from
// it. A privilege named in both is removed. The name and public key are unchanged.
type UpdateUserRequest struct {
	Name            string
	Description     string
	IsDescription   bool
	SetPrivileges   []string
	UnsetPrivileges []string
}

// MarshalCbor encodes the request as a CBOR array:
// [name, description, isDescription, setPrivileges, unsetPrivileges].
func (m *UpdateUserRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(5, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Name, m.Description} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	err = cboring.WriteBoolean(m.IsDescription, w)
	if err != nil {
		return err
	}

	err = writeStringSlice(m.SetPrivileges, w)
	if err != nil {
		return err
	}

	return writeStringSlice(m.UnsetPrivileges, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *UpdateUserRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 5 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Name, &m.Description} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	isDescription, err := cboring.ReadBoolean(r)
	if err != nil {
		return err
	}
	m.IsDescription = isDescription

	setPrivileges, err := readStringSlice(r)
	if err != nil {
		return err
	}
	m.SetPrivileges = setPrivileges

	unsetPrivileges, err := readStringSlice(r)
	if err != nil {
		return err
	}
	m.UnsetPrivileges = unsetPrivileges

	return nil
}
