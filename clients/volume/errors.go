package volume

import (
	"errors"

	"github.com/nats-io/nats.go/jetstream"
)

var (
	ErrVolumeExists      = errors.New("volume already exists")
	ErrVolumeNotFound    = errors.New("volume not found")
	ErrFileNotFound      = errors.New("file not found")
	ErrInvalidVolumeName = errors.New("invalid volume name")
)

func translate(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, jetstream.ErrBucketExists):
		return ErrVolumeExists
	case errors.Is(err, jetstream.ErrBucketNotFound):
		return ErrVolumeNotFound
	case errors.Is(err, jetstream.ErrObjectNotFound):
		return ErrFileNotFound
	case errors.Is(err, jetstream.ErrInvalidStoreName):
		return ErrInvalidVolumeName
	default:
		return err
	}
}
