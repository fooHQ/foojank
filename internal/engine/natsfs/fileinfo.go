package natsfs

import (
	"time"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.FileInfo = &FileInfo{}

type FileInfo struct {
	name    string
	size    int64
	mode    risoros.FileMode
	modTime time.Time
}

func (f *FileInfo) Name() string {
	return f.name
}

func (f *FileInfo) Size() int64 {
	return f.size
}

func (f *FileInfo) Mode() risoros.FileMode {
	return f.mode
}

func (f *FileInfo) ModTime() time.Time {
	return f.modTime
}

func (f *FileInfo) IsDir() bool {
	return false
}

func (f *FileInfo) Sys() any {
	return nil
}
