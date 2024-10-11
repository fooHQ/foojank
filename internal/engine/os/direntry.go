package os

import (
	"io/fs"
)

type DirEntry struct {
	name string
	mode fs.FileMode
}

func (e *DirEntry) Name() string {
	//TODO implement me
	panic("implement me")
}

func (e *DirEntry) IsDir() bool {
	//TODO implement me
	panic("implement me")
}

func (e *DirEntry) Type() fs.FileMode {
	//TODO implement me
	panic("implement me")
}

func (e *DirEntry) Info() (fs.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (e *DirEntry) HasInfo() bool {
	//TODO implement me
	panic("implement me")
}
