package agent

import (
	"context"

	natsfs "github.com/foohq/ren-natsfs"
	"github.com/nats-io/nats.go/jetstream"
)

type Storage struct {
	*natsfs.FS
	store jetstream.ObjectStore
}

type StorageStatus struct {
	Name        string
	Description string
	Size        uint64
}

func NewStorage(ctx context.Context, store jetstream.ObjectStore) (*Storage, error) {
	fs, err := natsfs.NewFS(ctx, store)
	if err != nil {
		return nil, err
	}
	return &Storage{
		FS:    fs,
		store: store,
	}, nil
}

func (o *Storage) Status(ctx context.Context) (*StorageStatus, error) {
	status, err := o.store.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &StorageStatus{
		Name:        status.Bucket(),
		Description: status.Description(),
		Size:        status.Size(),
	}, nil
}

func (o *Storage) Close() error {
	err := o.FS.Close()
	if err != nil {
		return err
	}
	return nil
}
