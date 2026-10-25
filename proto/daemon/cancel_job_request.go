package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// CancelJobRequest is a request to cancel a job.
type CancelJobRequest struct {
	JobRef string
}

// MarshalCbor encodes the request as a CBOR array: [jobRef].
func (m *CancelJobRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.JobRef, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *CancelJobRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	jobRef, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.JobRef = jobRef

	return nil
}
