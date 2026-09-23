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
	Config      GatewayConfig
}

// MarshalCbor encodes the gateway as a CBOR array:
// [id, name, description, config].
func (m *Gateway) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.ID, m.Name, m.Description} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return m.Config.MarshalCbor(w)
}

// UnmarshalCbor decodes the gateway from its CBOR representation.
func (m *Gateway) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.ID, &m.Name, &m.Description} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	var config GatewayConfig
	err = config.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.Config = config

	return nil
}
