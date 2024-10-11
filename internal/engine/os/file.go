package os

import (
	risoros "github.com/risor-io/risor/os"
	"io"
	"io/fs"
	"time"
)

// TODO: Rename to Pipe!

var _ risoros.File = File{}

type File struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func NewFile() File {
	r, w := io.Pipe()
	return File{
		r: r,
		w: w,
	}
}

func (f File) Write(p []byte) (n int, err error) {
	return f.w.Write(p)
}

func (f File) Stat() (fs.FileInfo, error) {
	return risoros.NewFileInfo(risoros.GenericFileInfoOpts{
		Name:    "grr",
		Size:    0,
		Mode:    0,
		ModTime: time.Time{},
		IsDir:   false,
	}), nil
}

func (f File) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

func (f File) Close() error {
	err := f.w.Close()
	if err != nil {
		return err
	}

	err = f.r.Close()
	if err != nil {
		return err
	}

	return nil
}
