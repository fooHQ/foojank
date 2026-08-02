package agent

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// AgentInfo is an event carrying information about an agent.
type AgentInfo struct {
	Username string
	Hostname string
	System   string
	Address  string
}

// MarshalCbor encodes the info as a CBOR array: [username, hostname, system, address].
func (m *AgentInfo) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Username, m.Hostname, m.System, m.Address} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalCbor decodes the info from its CBOR representation.
func (m *AgentInfo) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Username, &m.Hostname, &m.System, &m.Address} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	return nil
}
