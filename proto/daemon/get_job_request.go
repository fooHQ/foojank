package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetJobRequest is a request to retrieve a job.
type GetJobRequest struct {
	JobID string
}

// MarshalCbor encodes the request as a CBOR array: [jobID].
func (m *GetJobRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.JobID, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *GetJobRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	jobID, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.JobID = jobID

	return nil
}
