package os

import (
	"errors"
	"fmt"

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
	file, err := fs.Create(pth)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	return file, nil
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
	err = fs.Mkdir(pth, perm)
	if err != nil {
		return fmt.Errorf("mkdir %s: %w", name, err)
	}
	return nil
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
	err = fs.MkdirAll(pth, perm)
	if err != nil {
		return fmt.Errorf("mkdir %s: %w", path, err)
	}
	return nil
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
	file, err := fs.Open(pth)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	return file, nil
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
	file, err := fs.OpenFile(pth, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	return file, nil
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
	b, err := fs.ReadFile(pth)
	if err != nil {
		return nil, fmt.Errorf("open %s: %v", name, err)
	}
	return b, nil
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
	err = fs.Remove(pth)
	if err != nil {
		return fmt.Errorf("remove %s: %w", name, err)
	}
	return nil
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
	err = fs.RemoveAll(pth)
	if err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
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
	err = oldFS.Rename(oldPth, newPth)
	if err != nil {
		return fmt.Errorf("rename %s %s: %w", oldPath, newPath, err)
	}
	return nil
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
	info, err := fs.Stat(pth)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", name, err)
	}
	return info, nil
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
	err = oldFS.Symlink(oldPth, newPth)
	if err != nil {
		return fmt.Errorf("symlink %s %s: %w", oldPth, newPth, err)
	}
	return nil
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
	err = fs.WriteFile(pth, data, perm)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	return nil
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
		return nil, fmt.Errorf("open %s: %w", name, err)
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
