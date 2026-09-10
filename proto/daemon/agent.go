package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// Agent is an agent as returned by the daemon.
type Agent struct {
	ID          string
	Name        string
	Description string
	GatewayID   string
	Config      AgentConfig
	CreatedAt   int64
}

// MarshalCbor encodes the agent as a CBOR array:
// [id, name, description, gatewayID, config, createdAt].
func (m *Agent) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(6, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.ID, m.Name, m.Description, m.GatewayID} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	err = m.Config.MarshalCbor(w)
	if err != nil {
		return err
	}

	return cboring.WriteInt(m.CreatedAt, w)
}

// UnmarshalCbor decodes the agent from its CBOR representation.
func (m *Agent) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 6 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.ID, &m.Name, &m.Description, &m.GatewayID} {
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

	createdAt, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}

	m.Config = config
	m.CreatedAt = createdAt

	return nil
}
