package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// ListGatewaysResponse is a response to a ListGatewaysRequest.
type ListGatewaysResponse struct {
	Gateways []Gateway
	Error    error
}

// MarshalCbor encodes the response as a CBOR array: [gateways, error].
func (m *ListGatewaysResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = writeGatewaySlice(m.Gateways, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *ListGatewaysResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	gateways, err := readGatewaySlice(r)
	if err != nil {
		return err
	}
	m.Gateways = gateways

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
