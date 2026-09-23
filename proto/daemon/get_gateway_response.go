package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetGatewayResponse is a response to a GetGatewayRequest.
type GetGatewayResponse struct {
	Gateway Gateway
	Error   error
}

// MarshalCbor encodes the response as a CBOR array: [gateway, error].
func (m *GetGatewayResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = m.Gateway.MarshalCbor(w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *GetGatewayResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	var gateway Gateway
	err = gateway.UnmarshalCbor(r)
	if err != nil {
		return err
	}
	m.Gateway = gateway

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
