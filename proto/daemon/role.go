package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// Role is a role as returned by the daemon.
type Role struct {
	Name        string
	Description string
	Privileges  []string
	CreatedAt   int64
}

// MarshalCbor encodes the role as a CBOR array:
// [name, description, privileges, createdAt].
func (m *Role) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Name, m.Description} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	err = writeStringSlice(m.Privileges, w)
	if err != nil {
		return err
	}

	return cboring.WriteInt(m.CreatedAt, w)
}

// UnmarshalCbor decodes the role from its CBOR representation.
func (m *Role) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Name, &m.Description} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	privileges, err := readStringSlice(r)
	if err != nil {
		return err
	}

	createdAt, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}

	m.Privileges = privileges
	m.CreatedAt = createdAt

	return nil
}
