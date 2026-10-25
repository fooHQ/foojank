package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// Job is a job as returned by the daemon.
type Job struct {
	ID          string
	GatewayID   string
	GatewayName string
	AgentID     string
	AgentName   string
	WorkerID    string
	Command     string
	Args        []string
	Env         []string
	Status      JobStatus
	CreatedAt   int64
}

// MarshalCbor encodes the job as a CBOR array:
// [id, gatewayID, gatewayName, agentID, agentName, workerID, command, args, env, status, createdAt].
func (m *Job) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(11, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.ID, m.GatewayID, m.GatewayName, m.AgentID, m.AgentName, m.WorkerID, m.Command} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	err = writeStringSlice(m.Args, w)
	if err != nil {
		return err
	}

	err = writeStringSlice(m.Env, w)
	if err != nil {
		return err
	}

	err = m.Status.MarshalCbor(w)
	if err != nil {
		return err
	}

	return cboring.WriteInt(m.CreatedAt, w)
}

// UnmarshalCbor decodes the job from its CBOR representation.
func (m *Job) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 11 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.ID, &m.GatewayID, &m.GatewayName, &m.AgentID, &m.AgentName, &m.WorkerID, &m.Command} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	args, err := readStringSlice(r)
	if err != nil {
		return err
	}

	env, err := readStringSlice(r)
	if err != nil {
		return err
	}

	var status JobStatus
	err = status.UnmarshalCbor(r)
	if err != nil {
		return err
	}

	createdAt, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}

	m.Args = args
	m.Env = env
	m.Status = status
	m.CreatedAt = createdAt

	return nil
}

// JobStatus is the latest reported status of a job.
// Error is stored status text, not an RPC failure.
type JobStatus struct {
	Status    string
	Error     string
	UpdatedAt int64
}

// MarshalCbor encodes the status as a CBOR array: [status, error, updatedAt].
func (m *JobStatus) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.Status, m.Error} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return cboring.WriteInt(m.UpdatedAt, w)
}

// UnmarshalCbor decodes the status from its CBOR representation.
func (m *JobStatus) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.Status, &m.Error} {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return err
		}
		*f = s
	}

	updatedAt, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}
	m.UpdatedAt = updatedAt

	return nil
}
