package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListJobsRequest is a request to list jobs.
// An empty AgentID lists every job.
type ListJobsRequest struct {
	AgentID string
}

// MarshalCbor encodes the request as a CBOR array: [agentID].
func (m *ListJobsRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.AgentID, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *ListJobsRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	agentID, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.AgentID = agentID

	return nil
}
