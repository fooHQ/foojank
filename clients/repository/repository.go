package repository

import (
	"github.com/nats-io/nats.go/jetstream"
	"io"
	"time"
)

type Repository struct {
	Name        string
	Description string
	Size        uint64
}

var _ io.Reader = &File{}

type File struct {
	object   jetstream.ObjectResult
	Name     string
	Size     uint64
	Modified time.Time
}

func (f *File) Read(b []byte) (int, error) {
	if f.object == nil {
		return 0, io.EOF
	}
	return f.object.Read(b)
}

func (f *File) Close() error {
	return f.object.Close()
}
