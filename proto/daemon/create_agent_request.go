package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CreateAgentRequest is a request to create an agent.
type CreateAgentRequest struct {
	Name        string
	Description string
	Gateway     string
	Config      AgentConfig
}

// MarshalCbor encodes the request as a CBOR array:
// [name, description, gateway, config].
func (m *CreateAgentRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Name, m.Description, m.Gateway} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return m.Config.MarshalCbor(w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CreateAgentRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Name, &m.Description, &m.Gateway} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	var config AgentConfig
	err = config.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.Config = config

	return nil
}

// AgentConfig is the build configuration for an agent.
type AgentConfig struct {
	OS    string
	Arch  string
	JWT   string
	Key   string
	Extra map[string]string
}

// MarshalCbor encodes the config as a CBOR array: [os, arch, jwt, key, extra].
func (m *AgentConfig) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(5, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.OS, m.Arch, m.JWT, m.Key} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return writeStringMap(m.Extra, w)
}

// UnmarshalCbor decodes the config from its CBOR representation.
func (m *AgentConfig) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 5 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.OS, &m.Arch, &m.JWT, &m.Key} {
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
