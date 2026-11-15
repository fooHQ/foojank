package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CreateGatewayRequest is a request to create a gateway.
type CreateGatewayRequest struct {
	Name        string
	Description string
	PublicKey   string
	Extra       map[string]string
}

// MarshalCbor encodes the request as a CBOR array:
// [name, description, publicKey, extra].
func (m *CreateGatewayRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Name, m.Description, m.PublicKey} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return writeStringMap(m.Extra, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CreateGatewayRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Name, &m.Description, &m.PublicKey} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	extra, err := readStringMap(r)
	if err != nil {
		return err
	}
	m.Extra = extra

	return nil
}
