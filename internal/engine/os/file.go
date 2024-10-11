package os

import (
	"bytes"
	"context"
	"io/fs"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
)

var _ risoros.File = &File{}

type File struct {
	name  string
	store jetstream.ObjectStore
}

func (f *File) Stat() (fs.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f *File) Read(b []byte) (int, error) {
	o, err := f.store.Get(context.TODO(), f.name)
	if err != nil {
		return 0, err
	}
	defer o.Close()

	return o.Read(b)
}

func (f *File) Close() error {
	//TODO implement me
	panic("implement me")
}

func (f *File) Write(b []byte) (int, error) {
	r := bytes.NewReader(b)
	_, err := f.store.Put(context.TODO(), jetstream.ObjectMeta{
		Name: f.name,
	}, r)
	if err != nil {
		return 0, err
	}

	return int(r.Size()), nil
}
