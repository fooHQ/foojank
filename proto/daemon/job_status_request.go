package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// JobStatusRequest reports a job's process exit status.
// Status is the process exit code.
type JobStatusRequest struct {
	JobRef string
	Status int64
	Error  error
}

// MarshalCbor encodes the request as a CBOR array: [jobRef, status, error].
func (m *JobStatusRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(3, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.JobRef, w)
	if err != nil {
		return err
	}

	err = cboring.WriteInt(m.Status, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *JobStatusRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 3 {
		return errors.New("invalid message array length")
	}

	jobRef, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.JobRef = jobRef

	status, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}
	m.Status = status

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
