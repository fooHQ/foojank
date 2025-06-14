package os

import (
	"errors"

	"github.com/foohq/urlpath"
	risoros "github.com/risor-io/risor/os"
)

var (
	ErrFSNotFound           = errors.New("filesystem not found")
	ErrCrossingFSBoundaries = errors.New("crossing filesystem boundaries")
)

var _ risoros.FS = &FS{}

type FS struct {
	registry map[string]risoros.FS
}

func NewFS() *FS {
	return &FS{
		registry: make(map[string]risoros.FS),
	}
}

func (f *FS) Create(name string) (risoros.File, error) {
	fs, err := f.lookupFS(name)
	if err != nil {
		return nil, err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return nil, err
	}
	return fs.Create(pth)
}

func (f *FS) Mkdir(name string, perm risoros.FileMode) error {
	fs, err := f.lookupFS(name)
	if err != nil {
		return err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return err
	}
	return fs.Mkdir(pth, perm)
}

func (f *FS) MkdirAll(path string, perm risoros.FileMode) error {
	fs, err := f.lookupFS(path)
	if err != nil {
		return err
	}
	pth, err := urlpath.Path(path)
	if err != nil {
		return err
	}
	return fs.MkdirAll(pth, perm)
}

func (f *FS) Open(name string) (risoros.File, error) {
	fs, err := f.lookupFS(name)
	if err != nil {
		return nil, err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return nil, err
	}
	return fs.Open(pth)
}

func (f *FS) OpenFile(name string, flag int, perm risoros.FileMode) (risoros.File, error) {
	fs, err := f.lookupFS(name)
	if err != nil {
		return nil, err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return nil, err
	}
	return fs.OpenFile(pth, flag, perm)
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	fs, err := f.lookupFS(name)
	if err != nil {
		return nil, err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(pth)
}

func (f *FS) Remove(name string) error {
	fs, err := f.lookupFS(name)
	if err != nil {
		return err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return err
	}
	return fs.Remove(pth)
}

func (f *FS) RemoveAll(path string) error {
	fs, err := f.lookupFS(path)
	if err != nil {
		return err
	}
	pth, err := urlpath.Path(path)
	if err != nil {
		return err
	}
	return fs.RemoveAll(pth)
}

func (f *FS) Rename(oldPath, newPath string) error {
	oldFS, err := f.lookupFS(oldPath)
	if err != nil {
		return err
	}
	newFS, err := f.lookupFS(newPath)
	if err != nil {
		return err
	}
	if oldFS != newFS {
		return ErrCrossingFSBoundaries
	}
	oldPth, err := urlpath.Path(oldPath)
	if err != nil {
		return err
	}
	newPth, err := urlpath.Path(newPath)
	if err != nil {
		return err
	}
	return oldFS.Rename(oldPth, newPth)
}

func (f *FS) Stat(name string) (risoros.FileInfo, error) {
	fs, err := f.lookupFS(name)
	if err != nil {
		return nil, err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return nil, err
	}
	return fs.Stat(pth)
}

func (f *FS) Symlink(oldName, newName string) error {
	oldFS, err := f.lookupFS(oldName)
	if err != nil {
		return err
	}
	newFS, err := f.lookupFS(newName)
	if err != nil {
		return err
	}
	if oldFS != newFS {
		return ErrCrossingFSBoundaries
	}
	oldPth, err := urlpath.Path(oldName)
	if err != nil {
		return err
	}
	newPth, err := urlpath.Path(newName)
	if err != nil {
		return err
	}
	return oldFS.Symlink(oldPth, newPth)
}

func (f *FS) WriteFile(name string, data []byte, perm risoros.FileMode) error {
	fs, err := f.lookupFS(name)
	if err != nil {
		return err
	}
	pth, err := urlpath.Path(name)
	if err != nil {
		return err
	}
	return fs.WriteFile(pth, data, perm)
}

func (f *FS) ReadDir(name string) ([]risoros.DirEntry, error) {
	fs, err := f.lookupFS(name)
	if err != nil {
		return nil, err
	}

	pth, err := urlpath.Path(name)
	if err != nil {
		return nil, err
	}

	results, err := fs.ReadDir(pth)
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
	fs, err := f.lookupFS(root)
	if err != nil {
		return err
	}
	pth, err := urlpath.Path(root)
	if err != nil {
		return err
	}
	return fs.WalkDir(pth, fn)
}

func (f *FS) lookupFS(pth string) (risoros.FS, error) {
	scheme, err := urlpath.Scheme(pth)
	if err != nil {
		return nil, err
	}

	if scheme == "" {
		scheme = "file"
	}

	fs, ok := f.registry[scheme]
	if !ok {
		return nil, ErrFSNotFound
	}
	return fs, nil
}
