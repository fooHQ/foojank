package directory

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var (
	ErrStorageExists      = errors.New("storage already exists")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrKeyNotFound        = errors.New("key not found")
	ErrKeyExists          = errors.New("key already exists")
	ErrNameInvalid        = errors.New("invalid name")
)

func translate(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, jetstream.ErrBucketExists):
		return ErrStorageExists
	case errors.Is(err, nats.ErrNoResponders), errors.Is(err, jetstream.ErrNoStreamResponse):
		return ErrServiceUnavailable
	case errors.Is(err, jetstream.ErrKeyNotFound):
		return ErrKeyNotFound
	case errors.Is(err, jetstream.ErrKeyExists):
		return ErrKeyExists
	case errors.Is(err, jetstream.ErrInvalidKey):
		return ErrNameInvalid
	default:
		return err
	}
}
