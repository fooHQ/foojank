package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetJobOutputRequest is a request for the output of a job.
// Offset is the number of bytes the caller has already read. Zero starts at
// the beginning.
type GetJobOutputRequest struct {
	JobID  string
	Offset uint64
}

// MarshalCbor encodes the request as a CBOR array: [jobID, offset].
func (m *GetJobOutputRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.JobID, w)
	if err != nil {
		return err
	}

	return cboring.WriteUInt(m.Offset, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *GetJobOutputRequest) UnmarshalCbor(r io.Reader) error {
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

	offset, err := cboring.ReadUInt(r)
	if err != nil {
		return err
	}
	m.Offset = offset

	return nil
}
