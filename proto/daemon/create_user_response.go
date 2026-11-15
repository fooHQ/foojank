package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CreateUserResponse is a response to a CreateUserRequest.
// ID is the identifier assigned to the created user.
type CreateUserResponse struct {
	ID    string
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [id, error].
func (m *CreateUserResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.ID, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *CreateUserResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	id, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.ID = id

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
