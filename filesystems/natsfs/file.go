package natsfs

import (
	"bytes"
	"context"
	"errors"
	"os"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
)

var _ risoros.File = &File{}

type File struct {
	name  string
	store jetstream.ObjectStore
}

func NewFile(name string, store jetstream.ObjectStore) *File {
	return &File{
		name:  name,
		store: store,
	}
}

func (f *File) Stat() (risoros.FileInfo, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (f *File) Read(b []byte) (int, error) {
	o, err := f.store.Get(context.TODO(), f.name)
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			return 0, os.ErrNotExist
		}
		return 0, err
	}
	return o.Read(b)
}

func (f *File) Close() error {
	return nil
}

func (f *File) Write(b []byte) (int, error) {
	object, err := f.store.Put(context.TODO(), jetstream.ObjectMeta{
		Name: f.name,
	}, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}

	return int(object.Size), nil
}
