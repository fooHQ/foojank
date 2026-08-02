package gateway

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// RegisterAgentResponse is a response to a RegisterAgentRequest. Env holds the
// gateway-provided part of the environment, such as API tokens and URLs, which
// is merged with the environment from the request.
type RegisterAgentResponse struct {
	Env   map[string]string
	Error error
}

// MarshalCbor encodes the response as a CBOR array: [env, error].
func (m *RegisterAgentResponse) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(2, w)
	if err != nil {
		return err
	}

	err = writeStringMap(m.Env, w)
	if err != nil {
		return err
	}

	return writeError(m.Error, w)
}

// UnmarshalCbor decodes the response from its CBOR representation.
func (m *RegisterAgentResponse) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 2 {
		return errors.New("invalid message array length")
	}

	env, err := readStringMap(r)
	if err != nil {
		return err
	}
	m.Env = env

	respErr, err := readError(r)
	if err != nil {
		return err
	}
	m.Error = respErr

	return nil
}
