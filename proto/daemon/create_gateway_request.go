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
	Config      GatewayConfig
}

// MarshalCbor encodes the request as a CBOR array:
// [name, description, config].
func (m *CreateGatewayRequest) MarshalCbor(w io.Writer) error {
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

	return m.Config.MarshalCbor(w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CreateGatewayRequest) UnmarshalCbor(r io.Reader) error {
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

	var config GatewayConfig
	err = config.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.Config = config

	return nil
}

// GatewayConfig is the configuration for a gateway.
type GatewayConfig struct {
	JWT   string
	Key   string
	Extra map[string]string
}

// MarshalCbor encodes the config as a CBOR array: [jwt, key, extra].
func (m *GatewayConfig) MarshalCbor(w io.Writer) error {
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

	return writeStringMap(m.Extra, w)
}

// UnmarshalCbor decodes the config from its CBOR representation.
func (m *GatewayConfig) UnmarshalCbor(r io.Reader) error {
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

	extra, err := readStringMap(r)
	if err != nil {
		return err
	}
	m.Extra = extra

	return nil
}
