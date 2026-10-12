package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetJobResponse is a response to a GetJobRequest.
type GetJobResponse struct {
	Job   Job
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [job, error].
func (m *GetJobResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = m.Job.MarshalCbor(w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *GetJobResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	var job Job
	err = job.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.Job = job

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
