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
	GatewayRef  string
	OS          string
	Arch        string
	Extra       map[string]string
	Commands    map[string]AgentCommand
}

// MarshalCbor encodes the request as a CBOR array:
// [name, description, gatewayRef, os, arch, extra, commands].
func (m *CreateAgentRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(7, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Name, m.Description, m.GatewayRef, m.OS, m.Arch} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	err = writeStringMap(m.Extra, w)
	if err != nil {
		return err
	}

	return writeSortedMap(m.Commands, (*AgentCommand).MarshalCbor, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CreateAgentRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 7 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Name, &m.Description, &m.GatewayRef, &m.OS, &m.Arch} {
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

	commands, err := readSortedMap((*AgentCommand).UnmarshalCbor, r)
	if err != nil {
		return err
	}

	m.Extra = extra
	m.Commands = commands

	return nil
}
