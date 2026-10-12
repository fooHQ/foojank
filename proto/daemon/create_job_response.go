package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CreateJobResponse is a response to a CreateJobRequest.
// JobID is the identifier assigned to the created job.
type CreateJobResponse struct {
	JobID string
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [jobID, error].
func (m *CreateJobResponse) MarshalCbor(w io.Writer) error {
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

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *CreateJobResponse) UnmarshalCbor(r io.Reader) error {
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
