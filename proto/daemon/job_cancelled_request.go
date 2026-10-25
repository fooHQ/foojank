package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// JobCancelledRequest reports the outcome of cancelling a job.
// A nil Error means the cancel command succeeded. It does not mean the job
// has finished.
type JobCancelledRequest struct {
	JobRef string
	Error  error
}

// MarshalCbor encodes the request as a CBOR array: [jobRef, error].
func (m *JobCancelledRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.JobRef, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *JobCancelledRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	jobRef, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.JobRef = jobRef

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
