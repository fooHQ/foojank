package file

import (
	"errors"
	"os"
	"path/filepath"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.FS = &FS{}

type FS struct{}

func NewFS() (*FS, error) {
	return &FS{}, nil
}

func (f *FS) Create(name string) (risoros.File, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return file, nil
}

func (f *FS) Mkdir(name string, perm risoros.FileMode) error {
	err := os.Mkdir(name, perm)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func (f *FS) MkdirAll(path string, perm risoros.FileMode) error {
	err := os.MkdirAll(path, perm)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func (f *FS) Open(name string) (risoros.File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return file, nil
}

func (f *FS) OpenFile(name string, flag int, perm risoros.FileMode) (risoros.File, error) {
	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return file, nil
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return b, nil
}

func (f *FS) Remove(name string) error {
	err := os.Remove(name)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func (f *FS) RemoveAll(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func (f *FS) Rename(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func (f *FS) Stat(name string) (risoros.FileInfo, error) {
	info, err := os.Stat(name)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return info, nil
}

func (f *FS) Symlink(oldName, newName string) error {
	err := os.Symlink(oldName, newName)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func (f *FS) WriteFile(name string, data []byte, perm risoros.FileMode) error {
	err := os.WriteFile(name, data, perm)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func (f *FS) ReadDir(name string) ([]risoros.DirEntry, error) {
	results, err := os.ReadDir(name)
	if err != nil {
		return nil, errors.Unwrap(err)
	}

	entries := make([]risoros.DirEntry, 0, len(results))
	for _, result := range results {
		entries = append(entries, &risoros.DirEntryWrapper{
			DirEntry: result,
		})
	}

	return entries, nil
}

func (f *FS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}
