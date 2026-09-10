package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CreateUserResponse is a response to a CreateUserRequest.
type CreateUserResponse struct {
	JWT   string
	Key   string
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [jwt, key, error].
func (m *CreateUserResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.JWT, m.Key} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *CreateUserResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.JWT, &m.Key} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
