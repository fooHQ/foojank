package daemon

import (
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

// Event is an event as returned by the daemon.
type Event struct{}

// MarshalCbor encodes the event as an empty CBOR array.
func (m *Event) MarshalCbor(w io.Writer) error {
	return cboring.WriteArrayLength(0, w)
}

// UnmarshalCbor decodes the event from its CBOR representation.
func (m *Event) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 0 {
		return errors.New("invalid message array length")
	}

	return nil
}
