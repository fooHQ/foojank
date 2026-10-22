package daemon

import (
	"errors"
	"io"
	"sort"

	"github.com/dtn7/cboring"
)

// Agent is an agent as returned by the daemon.
type Agent struct {
	ID          string
	Name        string
	Description string
	GatewayID   string
	Config      AgentConfig
	Commands    map[string]AgentCommand
	CreatedAt   int64
}

// MarshalCbor encodes the agent as a CBOR array:
// [id, name, description, gatewayID, config, commands, createdAt].
func (m *Agent) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(7, w)
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

	err = writeSortedMap(m.Commands, (*AgentCommand).MarshalCbor, w)
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
	if l != 7 {
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

	commands, err := readSortedMap((*AgentCommand).UnmarshalCbor, r)
	if err != nil {
		return err
	}

	createdAt, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}

	m.Config = config
	m.Commands = commands
	m.CreatedAt = createdAt

	return nil
}

// AgentCommand is a command an agent exposes.
type AgentCommand struct {
	Description string
	Flags       map[string]AgentCommandFlag
	Environment map[string]AgentCommandVar
}

// MarshalCbor encodes the command as a CBOR array:
// [description, flags, environment].
func (m *AgentCommand) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.Description, w)
	if err != nil {
		return err
	}

	err = writeSortedMap(m.Flags, (*AgentCommandFlag).MarshalCbor, w)
	if err != nil {
		return err
	}

	return writeSortedMap(m.Environment, (*AgentCommandVar).MarshalCbor, w)
}

// UnmarshalCbor decodes the command from its CBOR representation.
func (m *AgentCommand) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
		return errors.New("invalid message array length")
	}

	description, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}

	flags, err := readSortedMap((*AgentCommandFlag).UnmarshalCbor, r)
	if err != nil {
		return err
	}

	environment, err := readSortedMap((*AgentCommandVar).UnmarshalCbor, r)
	if err != nil {
		return err
	}

	m.Description = description
	m.Flags = flags
	m.Environment = environment

	return nil
}

// AgentCommandFlag is a flag accepted by an agent command.
type AgentCommandFlag struct {
	Description string
	Default     string
	Required    bool
}

// MarshalCbor encodes the flag as a CBOR array: [description, default, required].
func (m *AgentCommandFlag) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Description, m.Default} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return cboring.WriteBoolean(m.Required, w)
}

// UnmarshalCbor decodes the flag from its CBOR representation.
func (m *AgentCommandFlag) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Description, &m.Default} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	required, err := cboring.ReadBoolean(r)
	if err != nil {
		return err
	}
	m.Required = required

	return nil
}

// AgentCommandVar is an environment variable for an agent command.
type AgentCommandVar struct {
	Value       string
	Description string
}

// MarshalCbor encodes the variable as a CBOR array: [value, description].
func (m *AgentCommandVar) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Value, m.Description} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalCbor decodes the variable from its CBOR representation.
func (m *AgentCommandVar) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Value, &m.Description} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	return nil
}

// writeSortedMap writes a map as a definite-length CBOR map. Keys are written
// in sorted order for deterministic output. A nil or empty map is an empty map.
func writeSortedMap[T any](m map[string]T, write func(*T, io.Writer) error, w io.Writer) error {
	err := cboring.WriteMapPairLength(uint64(len(m)), w)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := cboring.WriteTextString(k, w)
		if err != nil {
			return err
		}

		v := m[k]
		err = write(&v, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// readSortedMap reads a definite-length CBOR map. An empty map is decoded as nil.
func readSortedMap[T any](read func(*T, io.Reader) error, r io.Reader) (map[string]T, error) {
	l, err := cboring.ReadMapPairLength(r)
	if err != nil {
		return nil, err
	}

	if l == 0 {
		return nil, nil
	}

	m := make(map[string]T, l)
	for range l {
		k, err := cboring.ReadTextString(r)
		if err != nil {
			return nil, err
		}

		var v T
		err = read(&v, r)
		if err != nil {
			return nil, err
		}
		m[k] = v
	}

	return m, nil
}
