package agent

import (
	"errors"

	"github.com/nats-io/nats.go/jetstream"
)

type errorApi struct {
	err error
}

func (e *errorApi) Error() string {
	switch {
	case errors.Is(e.err, jetstream.ErrStreamNotFound):
		return "agent not found"
	}
	return e.err.Error()
}
