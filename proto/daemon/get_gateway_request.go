package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// GetGatewayRequest is a request to retrieve a gateway.
// GatewayRef identifies the gateway.
type GetGatewayRequest struct {
	GatewayRef string
}

// MarshalCbor encodes the request as a CBOR array: [gatewayRef].
func (m *GetGatewayRequest) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(1, w)
	if err != nil {
		return err
	}

	return cboring.WriteTextString(m.GatewayRef, w)
}

// UnmarshalCbor decodes the request from its CBOR representation.
func (m *GetGatewayRequest) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("invalid message array length")
	}

	gatewayRef, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}
	m.GatewayRef = gatewayRef

	return nil
}
