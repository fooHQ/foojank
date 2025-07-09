package repository

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	natsfs "github.com/foohq/ren/filesystems/nats"
)

type Repository struct {
	*natsfs.FS
	name        string
	description string
	size        uint64
}

func New(ctx context.Context, store jetstream.ObjectStore) (*Repository, error) {
	fs, err := natsfs.NewFS(ctx, store)
	if err != nil {
		return nil, err
	}

	status, err := store.Status(ctx)
	if err != nil {
		return nil, err
	}

	return &Repository{
		FS:          fs,
		name:        status.Bucket(),
		description: status.Description(),
		size:        status.Size(),
	}, nil
}

func (r *Repository) Name() string {
	return r.name
}

func (r *Repository) Description() string {
	return r.description
}

func (r *Repository) Size() uint64 {
	return r.size
}

func (r *Repository) Close() error {
	err := r.FS.Close()
	if err != nil {
		return err
	}
	return nil
}
