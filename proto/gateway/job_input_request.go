package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// JobInputRequest is a request to pass one chunk of input to a job.
type JobInputRequest struct {
	AgentID string
	JobID   string
	Data    []byte
}

// MarshalCbor encodes the request as a CBOR array: [agentID, jobID, data].
func (m *JobInputRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.AgentID, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.JobID, w)
	if err != nil {
		return err
	}

	return cboring.WriteByteString(m.Data, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *JobInputRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
		return errors.New("invalid message array length")
	}

	agentID, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.AgentID = agentID

	jobID, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.JobID = jobID

	data, err := cboring.ReadByteString(r)
	if err != nil {
		return err
	}
	m.Data = data

	return nil
}
