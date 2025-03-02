package nats

import (
	risoros "github.com/risor-io/risor/os"
)

var _ risoros.DirEntry = &DirEntry{}

type DirEntry struct {
	name string
	mode risoros.FileMode
}

func (e *DirEntry) Name() string {
	return e.name
}

func (e *DirEntry) IsDir() bool {
	return e.mode.IsDir()
}

func (e *DirEntry) Type() risoros.FileMode {
	return e.mode
}

func (e *DirEntry) Info() (risoros.FileInfo, error) {
	return nil, ErrUnsupportedOperation
}

func (e *DirEntry) HasInfo() bool {
	return false
}
