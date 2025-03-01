package filefs

import (
	"os"
	"path/filepath"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.FS = &FS{}

type FS struct{}

func NewFS() *FS {
	return &FS{}
}

func (f *FS) Create(name string) (risoros.File, error) {
	return os.Create(name)
}

func (f *FS) Mkdir(name string, perm risoros.FileMode) error {
	return os.Mkdir(name, perm)
}

func (f *FS) MkdirAll(path string, perm risoros.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *FS) Open(name string) (risoros.File, error) {
	return os.Open(name)
}

func (f *FS) OpenFile(name string, flag int, perm risoros.FileMode) (risoros.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (f *FS) Remove(name string) error {
	return os.Remove(name)
}

func (f *FS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (f *FS) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (f *FS) Stat(name string) (risoros.FileInfo, error) {
	return os.Stat(name)
}

func (f *FS) Symlink(oldName, newName string) error {
	return os.Symlink(oldName, newName)
}

func (f *FS) WriteFile(name string, data []byte, perm risoros.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (f *FS) ReadDir(name string) ([]risoros.DirEntry, error) {
	results, err := os.ReadDir(name)
	if err != nil {
		return nil, err
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
