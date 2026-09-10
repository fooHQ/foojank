package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// IssueJWTResponse is a response to an IssueJWTRequest.
type IssueJWTResponse struct {
	JWT   string
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [jwt, error].
func (m *IssueJWTResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.JWT, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *IssueJWTResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	jwt, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.JWT = jwt

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
