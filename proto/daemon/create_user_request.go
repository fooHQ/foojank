package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CreateUserRequest is a request to create a user.
type CreateUserRequest struct {
	Name        string
	Description string
	Privileges  []string
}

// MarshalCbor encodes the request as a CBOR array: [name, description, privileges].
func (m *CreateUserRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Name, m.Description} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return writeStringSlice(m.Privileges, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CreateUserRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
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
	m.Privileges = privileges

	return nil
}
