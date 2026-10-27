package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// HealthCheckResponse is a response to a HealthCheckRequest.
type HealthCheckResponse struct {
	Status string
	Error  error
}

// MarshalCbor encodes the response as a CBOR array: [status, error].
func (m *HealthCheckResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.Status, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *HealthCheckResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	status, err := cboring.ReadTextString(r)
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
