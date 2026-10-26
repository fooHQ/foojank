package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListAuditRequest is a request to list audit records.
// Offset is the number of records to skip. Limit is the maximum number of
// records to return.
type ListAuditRequest struct {
	Offset int64
	Limit  int64
}

// MarshalCbor encodes the request as a CBOR array: [offset, limit].
func (m *ListAuditRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteInt(m.Offset, w)
	if err != nil {
		return err
	}

	return cboring.WriteInt(m.Limit, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *ListAuditRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	offset, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}
	m.Offset = offset

	limit, err := cboring.ReadInt(r)
	if err != nil {
		return err
	}
	m.Limit = limit

	return nil
}
