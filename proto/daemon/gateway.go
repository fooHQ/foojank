package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// Gateway is a gateway as returned by the daemon.
type Gateway struct {
	ID          string
	Name        string
	Description string
	JWT         string
	Extra       map[string]string
}

// MarshalCbor encodes the gateway as a CBOR array:
// [id, name, description, jwt, extra].
func (m *Gateway) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(5, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.ID, m.Name, m.Description, m.JWT} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return writeStringMap(m.Extra, w)
}

// UnmarshalCbor decodes the gateway from its CBOR representation.
func (m *Gateway) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 5 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.ID, &m.Name, &m.Description, &m.JWT} {
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
