package natsfs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
)

var (
	ErrInvalidFilename      = errors.New("invalid filename")
	ErrUnsupportedOperation = errors.New("unsupported operation")
)

var _ risoros.FS = &FS{}

type FS struct {
	store jetstream.ObjectStore
}

// TODO: context should have a timeout!

func New(store jetstream.ObjectStore) (*FS, error) {
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
	return ErrUnsupportedOperation
}

func (f *FS) MkdirAll(path string, perm risoros.FileMode) error {
	return ErrUnsupportedOperation
}

func (f *FS) Open(name string) (risoros.File, error) {
	_, err := f.store.GetInfo(context.TODO(), name)
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			return nil, os.ErrNotExist
		}
		return nil, err
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
	info, err := f.store.GetInfo(context.TODO(), name)
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	return &FileInfo{
		name:    info.Name,
		size:    int64(info.Size),
		mode:    0777,
		modTime: info.ModTime,
	}, nil
}

func (f *FS) Symlink(oldName, newName string) error {
	// This operation is supported by ObjectStore, however at this point there seems to be
	// no actual use case.
	return ErrUnsupportedOperation
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
	if name == "" {
		return nil, ErrInvalidFilename
	}

	files, err := f.store.List(context.TODO())
	if err != nil {
		if errors.Is(err, jetstream.ErrNoObjectsFound) {
			return nil, nil
		}
		return nil, err
	}

	entries := matchObjects(files, name)
	return entries, nil
}

func (f *FS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	return errors.New("unsupported")
}

func matchObjects(objects []*jetstream.ObjectInfo, path string) []risoros.DirEntry {
	var matched []risoros.DirEntry
	for _, object := range objects {
		// Check if the file path starts with the mask
		if strings.HasPrefix(object.Name, path) {
			// Remove the mask from the file path
			relativePath := strings.TrimPrefix(object.Name, path)
			// If the relative path starts with a path separator, remove it
			if strings.HasPrefix(relativePath, "/") {
				relativePath = relativePath[1:]
			}
			// Get the file name from the relative path
			filename := filepath.Base(filepath.Dir(relativePath))
			if filename == "." {
				filename = filepath.Base(relativePath)
			}
			// If the dir name is not empty, add it to the matched paths
			if filename != "" {
				var mode risoros.FileMode
				if filename != relativePath {
					mode = 0755 | risoros.FileMode(os.ModeDir)
				} else {
					mode = 0644
				}
				matched = append(matched, &DirEntry{
					name: filename,
					mode: mode,
				})
			}
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Name() < matched[j].Name()
	})

	matched = slices.CompactFunc(matched, func(i, j risoros.DirEntry) bool {
		return i.Name() == j.Name()
	})

	return matched
}
