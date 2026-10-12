package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CancelJobRequest is a request to cancel a job on the gateway.
type CancelJobRequest struct {
	AgentID string
	JobID   string
}

// MarshalCbor encodes the request as a CBOR array: [agentID, jobID].
func (m *CancelJobRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.AgentID, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.JobID, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CancelJobRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
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

	return nil
}
