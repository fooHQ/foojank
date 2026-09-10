package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// User is a NATS user as returned by the daemon.
type User struct {
	ID          string
	Name        string
	Description string
	Kind        string
	IssuedAt    int64
	ExpiresAt   int64
}

// MarshalCbor encodes the user as a CBOR array:
// [name, id, description, kind, issuedAt, expiresAt].
func (m *User) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(6, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Name, m.ID, m.Description, m.Kind} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	for _, n := range []int64{m.IssuedAt, m.ExpiresAt} {
		err := cboring.WriteInt(n, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalCbor decodes the user from its CBOR representation.
func (m *User) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 6 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Name, &m.ID, &m.Description, &m.Kind} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	issuedAt, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}

	expiresAt, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}

	m.IssuedAt = issuedAt
	m.ExpiresAt = expiresAt

	return nil
}
