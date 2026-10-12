package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// JobStartedRequest reports the outcome of starting a job.
// A nil Error means the start succeeded.
type JobStartedRequest struct {
	JobID string
	Error error
}

// MarshalCbor encodes the request as a CBOR array: [jobID, error].
func (m *JobStartedRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.JobID, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *JobStartedRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	jobID, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.JobID = jobID

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
