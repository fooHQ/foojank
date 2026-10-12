package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListJobsResponse is a response to a ListJobsRequest.
type ListJobsResponse struct {
	Jobs  []Job
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [jobs, error].
func (m *ListJobsResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = writeJobSlice(m.Jobs, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *ListJobsResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	jobs, err := readJobSlice(r)
	if err != nil {
		return err
	}
	m.Jobs = jobs

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
