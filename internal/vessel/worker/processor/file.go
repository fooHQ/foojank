package processor

import (
	"bytes"
	"io"
	"time"
)

var (
	_ io.Reader   = &File{}
	_ io.ReaderAt = &File{}
)

type File struct {
	b        *bytes.Reader
	Name     string
	Size     uint64
	Modified time.Time
}

func (f *File) Read(b []byte) (int, error) {
	if f.b == nil {
		return 0, io.EOF
	}
	return f.b.Read(b)
}

func (f *File) ReadAt(b []byte, off int64) (int, error) {
	if f.b == nil {
		return 0, io.EOF
	}
	return f.b.ReadAt(b, off)
}
