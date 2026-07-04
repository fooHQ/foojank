package daemon

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrGatewayNotFound    = errors.New("gateway not found")
	ErrJobNotFound        = errors.New("job not found")
	ErrStorageExists      = errors.New("storage already exists")
	ErrStorageNotFound    = errors.New("storage not found")
	ErrFileNotFound       = errors.New("file not found")
	ErrInvalidStorageName = errors.New("invalid storage name")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrKeyNotFound        = errors.New("key not found")
	ErrKeyExists          = errors.New("key already exists")
	ErrKeyInvalid         = errors.New("invalid key")
	ErrNameInvalid        = errors.New("invalid name")
	ErrStreamNotFound     = errors.New("stream not found")
)

func translate(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, jetstream.ErrBucketExists):
		return ErrStorageExists
	case errors.Is(err, jetstream.ErrBucketNotFound):
		return ErrStorageNotFound
	case errors.Is(err, jetstream.ErrObjectNotFound):
		return ErrFileNotFound
	case errors.Is(err, jetstream.ErrInvalidStoreName):
		return ErrInvalidStorageName
	case errors.Is(err, nats.ErrNoResponders), errors.Is(err, jetstream.ErrNoStreamResponse):
		return ErrServiceUnavailable
	case errors.Is(err, jetstream.ErrKeyNotFound):
		return ErrKeyNotFound
	case errors.Is(err, jetstream.ErrKeyExists):
		return ErrKeyExists
	case errors.Is(err, jetstream.ErrInvalidKey):
		return ErrKeyInvalid
	case errors.Is(err, jetstream.ErrStreamNotFound):
		return ErrStreamNotFound
	default:
		return err
	}
}
