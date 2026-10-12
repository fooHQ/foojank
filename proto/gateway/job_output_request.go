package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// JobOutputRequest is a request for the output of a job.
// Offset is the number of bytes the caller has already read. Zero starts at
// the beginning.
type JobOutputRequest struct {
	AgentID string
	JobID   string
	Offset  uint64
}

// MarshalCbor encodes the request as a CBOR array: [agentID, jobID, offset].
func (m *JobOutputRequest) MarshalCbor(w io.Writer) error {
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

	return cboring.WriteUInt(m.Offset, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *JobOutputRequest) UnmarshalCbor(r io.Reader) error {
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

	offset, err := cboring.ReadUInt(r)
	if err != nil {
		return err
	}
	m.Offset = offset

	return nil
}
