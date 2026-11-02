package volume

import (
	"context"

	natsfs "github.com/foohq/ren-natsfs"
	"github.com/nats-io/nats.go/jetstream"
)

type Volume struct {
	*natsfs.FS
	name  string
	store jetstream.ObjectStore
}

type Status struct {
	Name        string
	Description string
	Size        uint64
}

func NewVolume(ctx context.Context, name string, store jetstream.ObjectStore) (*Volume, error) {
	fs, err := natsfs.NewFS(ctx, store)
	if err != nil {
		return nil, err
	}
	return &Volume{
		FS:    fs,
		name:  name,
		store: store,
	}, nil
}

func (o *Volume) Status(ctx context.Context) (*Status, error) {
	status, err := o.store.Status(ctx)
	if err != nil {
		return nil, err
	}
	name := o.name
	if name == "" {
		name = status.Bucket()
	}
	return &Status{
		Name:        name,
		Description: status.Description(),
		Size:        status.Size(),
	}, nil
}

func (o *Volume) Close() error {
	err := o.FS.Close()
	if err != nil {
		return err
	}
	return nil
}
