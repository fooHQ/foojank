package daemon

import (
	"errors"

	"github.com/nats-io/nats.go"
)

var (
	ErrDaemonUnavailable = errors.New("daemon is unavailable")
)

func translate(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, nats.ErrNoResponders):
		return ErrDaemonUnavailable
	default:
		return err
	}
}
