package testutils

import (
	risoros "github.com/risor-io/risor/os"
)

type FS struct {
	resultCh chan<- any
}

func NewFS(resultCh chan<- any) *FS {
	return &FS{
		resultCh: resultCh,
	}
}

type CreateResult struct {
	Name string
}

func (fs *FS) Create(name string) (risoros.File, error) {
	fs.resultCh <- CreateResult{
		Name: name,
	}
	return nil, nil
}

type MkdirResult struct {
	Name string
	Perm risoros.FileMode
}

func (fs *FS) Mkdir(name string, perm risoros.FileMode) error {
	fs.resultCh <- MkdirResult{
		Name: name,
		Perm: perm,
	}
	return nil
}

type MkdirAllResult struct {
	Path string
	Perm risoros.FileMode
}

func (fs *FS) MkdirAll(path string, perm risoros.FileMode) error {
	fs.resultCh <- MkdirAllResult{
		Path: path,
		Perm: perm,
	}
	return nil
}

type OpenResult struct {
	Name string
}

func (fs *FS) Open(name string) (risoros.File, error) {
	fs.resultCh <- OpenResult{
		Name: name,
	}
	return nil, nil
}

type OpenFileResult struct {
	Name string
	Flag int
	Perm risoros.FileMode
}

func (fs *FS) OpenFile(name string, flag int, perm risoros.FileMode) (risoros.File, error) {
	fs.resultCh <- OpenFileResult{
		Name: name,
		Flag: flag,
		Perm: perm,
	}
	return nil, nil
}

type ReadFileResult struct {
	Name string
}

func (fs *FS) ReadFile(name string) ([]byte, error) {
	fs.resultCh <- ReadFileResult{
		Name: name,
	}
	return nil, nil
}

type RemoveResult struct {
	Name string
}

func (fs *FS) Remove(name string) error {
	fs.resultCh <- RemoveResult{
		Name: name,
	}
	return nil
}

type RemoveAllResult struct {
	Path string
}

func (fs *FS) RemoveAll(path string) error {
	fs.resultCh <- RemoveAllResult{
		Path: path,
	}
	return nil
}

type RenameResult struct {
	OldPath string
	NewPath string
}

func (fs *FS) Rename(oldpath, newpath string) error {
	fs.resultCh <- RenameResult{
		OldPath: oldpath,
		NewPath: newpath,
	}
	return nil
}

type StatResult struct {
	Name string
}

func (fs *FS) Stat(name string) (risoros.FileInfo, error) {
	fs.resultCh <- StatResult{
		Name: name,
	}
	return nil, nil
}

type SymlinkResult struct {
	OldName string
	NewName string
}

func (fs *FS) Symlink(oldname, newname string) error {
	fs.resultCh <- SymlinkResult{
		OldName: oldname,
		NewName: newname,
	}
	return nil
}

type WriteFileResult struct {
	Name string
	Data []byte
	Perm risoros.FileMode
}

func (fs *FS) WriteFile(name string, data []byte, perm risoros.FileMode) error {
	fs.resultCh <- WriteFileResult{
		Name: name,
		Data: data,
		Perm: perm,
	}
	return nil
}

type ReadDirResult struct {
	Name string
}

func (fs *FS) ReadDir(name string) ([]risoros.DirEntry, error) {
	fs.resultCh <- ReadDirResult{
		Name: name,
	}
	return nil, nil
}

type WalkDirResult struct {
	Root string
	Fn   risoros.WalkDirFunc
}

func (fs *FS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	fs.resultCh <- WalkDirResult{
		Root: root,
		Fn:   fn,
	}
	return nil
}
