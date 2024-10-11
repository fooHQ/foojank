package os

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
)

var _ risoros.FS = &FS{}

type FS struct {
	js    jetstream.JetStream
	store jetstream.ObjectStore
}

// TODO: methods should return filepath errors, if possible!

func NewFS(js jetstream.JetStream, bucket string) (*FS, error) {
	s, err := js.CreateObjectStore(context.TODO(), jetstream.ObjectStoreConfig{
		Bucket: bucket,
		// TODO: add configurables
	})
	if err != nil {
		if !errors.Is(err, jetstream.ErrBucketExists) {
			return nil, err
		}
	}

	return &FS{
		js:    js,
		store: s,
	}, nil
}

func (f *FS) clean(name string) string {
	return filepath.Clean("/" + name)
}

func (f *FS) Create(name string) (risoros.File, error) {
	name = f.clean(name)
	_, err := f.store.Put(context.TODO(), jetstream.ObjectMeta{
		Name: name,
	}, strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	return &File{
		name:  name,
		store: f.store,
	}, nil
}

func (f *FS) Mkdir(name string, perm risoros.FileMode) error {
	return errors.New("unsupported")
}

func (f *FS) MkdirAll(path string, perm risoros.FileMode) error {
	return errors.New("unsupported")
}

func (f *FS) Open(name string) (risoros.File, error) {
	name = f.clean(name)
	_, err := f.store.GetInfo(context.TODO(), name)
	if err != nil {
		return nil, fs.ErrNotExist
	}

	return &File{
		name:  name,
		store: f.store,
	}, nil
}

func (f *FS) OpenFile(name string, flag int, perm risoros.FileMode) (risoros.File, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FS) Remove(name string) error {
	name = f.clean(name)
	err := f.store.Delete(context.TODO(), name)
	if errors.Is(err, jetstream.ErrObjectNotFound) {
		return fs.ErrNotExist
	}
	return err
}

func (f *FS) RemoveAll(path string) error {
	// If bucket exists delete it
	panic("implement me")
}

func (f *FS) Rename(oldpath, newpath string) error {
	// rename object
	panic("implement me")
}

func (f *FS) Stat(name string) (risoros.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FS) Symlink(oldname, newname string) error {
	//TODO implement me
	panic("implement me")
}

func (f *FS) WriteFile(name string, data []byte, perm risoros.FileMode) error {
	panic("implement me")
}

func (f *FS) ReadDir(name string) ([]risoros.DirEntry, error) {
	name = f.clean(name)
	l, err := f.store.List(context.TODO())
	if err != nil {
		return nil, err
	}

	var res []risoros.DirEntry
	for i := range l {
		cn := f.clean(l[i].Name)
		dir := filepath.Dir(cn)
		if name == dir {
			res = append(res, &DirEntry{
				name: cn,
				mode: 0,
			})
		}
	}
	// TODO:
	return nil, err
}

func (f *FS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	//TODO implement me
	panic("implement me")
}
