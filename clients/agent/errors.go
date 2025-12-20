package agent

import (
	"errors"

	"github.com/nats-io/nats.go/jetstream"
)

var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrStorageExists      = errors.New("storage already exists")
	ErrStorageNotFound    = errors.New("storage not found")
	ErrFileNotFound       = errors.New("file not found")
	ErrInvalidStorageName = errors.New("invalid storage name")
	ErrJobNotFound        = errors.New("job not found")
)

type errorApi struct {
	err error
}

func (e *errorApi) Error() string {
	switch {
	case errors.Is(e.err, jetstream.ErrStreamNotFound), errors.Is(e.err, jetstream.ErrNoStreamResponse):
		return ErrAgentNotFound.Error()
	case errors.Is(e.err, jetstream.ErrBucketExists):
		return ErrStorageExists.Error()
	case errors.Is(e.err, jetstream.ErrBucketNotFound):
		return ErrStorageNotFound.Error()
	case errors.Is(e.err, jetstream.ErrObjectNotFound):
		return ErrFileNotFound.Error()
	case errors.Is(e.err, jetstream.ErrInvalidStoreName):
		return ErrInvalidStorageName.Error()
	default:
		return e.err.Error()
	}
}
