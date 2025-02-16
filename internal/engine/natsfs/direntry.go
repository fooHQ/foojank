package natsfs

import (
	"io/fs"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.DirEntry = &DirEntry{}

type DirEntry struct {
	name string
	mode fs.FileMode
}

func (e *DirEntry) Name() string {
	return e.name
}

func (e *DirEntry) IsDir() bool {
	return false
}

func (e *DirEntry) Type() fs.FileMode {
	return e.mode
}

func (e *DirEntry) Info() (fs.FileInfo, error) {
	return nil, ErrUnsupportedOperation
}

func (e *DirEntry) HasInfo() bool {
	return false
}
