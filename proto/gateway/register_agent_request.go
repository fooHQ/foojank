package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// RegisterAgentRequest is a request to register an agent with the gateway.
// OS and Arch select the build target; Env holds the user-provided part of
// the environment the agent is compiled with.
type RegisterAgentRequest struct {
	AgentID string
	OS      string
	Arch    string
	Env     map[string]string
}

// MarshalCbor encodes the request as a CBOR array: [agentID, os, arch, env].
func (m *RegisterAgentRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.AgentID, m.OS, m.Arch} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return writeStringMap(m.Env, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *RegisterAgentRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.AgentID, &m.OS, &m.Arch} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	env, err := readStringMap(r)
	if err != nil {
		return err
	}
	m.Env = env

	return nil
}
