package agent

import (
	"context"

	natsfs "github.com/foohq/ren-natsfs"
	"github.com/nats-io/nats.go/jetstream"
)

type Storage struct {
	*natsfs.FS
	name        string
	description string
	size        uint64
}

func NewStorage(ctx context.Context, store jetstream.ObjectStore) (*Storage, error) {
	fs, err := natsfs.NewFS(ctx, store)
	if err != nil {
		return nil, err
	}

	status, err := store.Status(ctx)
	if err != nil {
		return nil, err
	}

	return &Storage{
		FS:          fs,
		name:        status.Bucket(),
		description: status.Description(),
		size:        status.Size(),
	}, nil
}

func (o *Storage) Name() string {
	return o.name
}

func (o *Storage) Description() string {
	return o.description
}

func (o *Storage) Size() uint64 {
	return o.size
}

func (o *Storage) Close() error {
	err := o.FS.Close()
	if err != nil {
		return err
	}
	return nil
}
