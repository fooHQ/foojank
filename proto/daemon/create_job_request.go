package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CreateJobRequest is a request to create a job.
// AgentRef is the agent that runs the job.
// Command, Args, and Env are the worker to start.
type CreateJobRequest struct {
	AgentRef string
	Command  string
	Args     []string
	Env      []string
}

// MarshalCbor encodes the request as a CBOR array:
// [agentRef, command, args, env].
func (m *CreateJobRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	for _, s := range []string{m.AgentRef, m.Command} {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	err = writeStringSlice(m.Args, w)
	if err != nil {
		return err
	}

	return writeStringSlice(m.Env, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CreateJobRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	for _, f := range []*string{&m.AgentRef, &m.Command} {
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

	m.Args = args
	m.Env = env

	return nil
}
