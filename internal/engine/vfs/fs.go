package vfs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
)

var _ risoros.FS = &FS{}

type FS struct {
	store jetstream.ObjectStore
}

// TODO: context should have a timeout!

func NewVirtualFS(store jetstream.ObjectStore) (*FS, error) {
	return &FS{
		store: store,
	}, nil
}

func (f *FS) Create(name string) (risoros.File, error) {
	_, err := f.store.Put(context.TODO(), jetstream.ObjectMeta{
		Name: name,
	}, nil)
	if err != nil {
		return nil, err
	}
	return f.Open(name)
}

func (f *FS) Mkdir(name string, perm risoros.FileMode) error {
	return errors.New("unsupported")
}

func (f *FS) MkdirAll(path string, perm risoros.FileMode) error {
	return errors.New("unsupported")
}

func (f *FS) Open(name string) (risoros.File, error) {
	_, err := f.store.GetInfo(context.TODO(), name)
	if err != nil {
		return nil, os.ErrNotExist
	}
	return &File{
		name:  name,
		store: f.store,
	}, nil
}

func (f *FS) OpenFile(name string, _ int, _ risoros.FileMode) (risoros.File, error) {
	return f.Open(name)
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	res, err := f.store.Get(context.TODO(), name)
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	defer res.Close()
	return io.ReadAll(res)
}

func (f *FS) Remove(name string) error {
	err := f.store.Delete(context.TODO(), name)
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			return os.ErrNotExist
		}
		return err
	}
	return nil
}

func (f *FS) RemoveAll(path string) error {
	return f.Remove(path)
}

func (f *FS) Rename(oldPath, newPath string) error {
	err := f.store.UpdateMeta(context.TODO(), oldPath, jetstream.ObjectMeta{
		Name: newPath,
	})
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			return os.ErrNotExist
		}
		return err
	}
	return nil
}

func (f *FS) Stat(name string) (risoros.FileInfo, error) {
	// TODO
	return nil, errors.New("unsupported")
}

func (f *FS) Symlink(oldName, newName string) error {
	// TODO
	return errors.New("unsupported")
}

func (f *FS) WriteFile(name string, data []byte, perm risoros.FileMode) error {
	_, err := f.store.Put(context.TODO(), jetstream.ObjectMeta{
		Name: name,
	}, bytes.NewReader(data))
	if err != nil {
		return err
	}
	return nil
}

func (f *FS) ReadDir(name string) ([]risoros.DirEntry, error) {
	// TODO
	return nil, errors.New("unsupported")
}

func (f *FS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	return errors.New("unsupported")
}
