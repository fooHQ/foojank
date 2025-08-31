package server

import (
	"context"
	"errors"

	natsfs "github.com/foohq/ren-natsfs"
	"github.com/nats-io/nats.go/jetstream"
)

type ObjectStore struct {
	*natsfs.FS
	name        string
	description string
	size        uint64
}

func NewObjectStore(ctx context.Context, store jetstream.ObjectStore) (*ObjectStore, error) {
	fs, err := natsfs.NewFS(ctx, store)
	if err != nil {
		return nil, err
	}

	status, err := store.Status(ctx)
	if err != nil {
		return nil, err
	}

	return &ObjectStore{
		FS:          fs,
		name:        status.Bucket(),
		description: status.Description(),
		size:        status.Size(),
	}, nil
}

func (o *ObjectStore) Name() string {
	return o.name
}

func (o *ObjectStore) Description() string {
	return o.description
}

func (o *ObjectStore) Size() uint64 {
	return o.size
}

func (o *ObjectStore) Close() error {
	err := o.FS.Close()
	if err != nil {
		return err
	}
	return nil
}

type errorObjectStore struct {
	err error
}

func (e *errorObjectStore) Error() string {
	switch {
	case errors.Is(e.err, jetstream.ErrBucketExists):
		return "repository already exists"
	case errors.Is(e.err, jetstream.ErrBucketNotFound):
		return "repository not found"
	case errors.Is(e.err, jetstream.ErrObjectNotFound):
		return "file not found"
	case errors.Is(e.err, jetstream.ErrInvalidStoreName):
		return "invalid repository name"
	}
	return e.err.Error()
}
