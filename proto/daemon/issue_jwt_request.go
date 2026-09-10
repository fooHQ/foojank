package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// IssueJWTRequest is a request to issue a JWT. The user is identified by the
// NATS subject, not by this payload.
type IssueJWTRequest struct{}

// MarshalCbor encodes the request as an empty CBOR array.
func (m *IssueJWTRequest) MarshalCbor(w io.Writer) error {
	return cboring.WriteArrayLength(0, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *IssueJWTRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 0 {
		return errors.New("invalid message array length")
	}

	return nil
}
