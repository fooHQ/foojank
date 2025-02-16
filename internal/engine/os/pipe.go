package os

import (
	"io"
	"time"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.File = &Pipe{}

// Pipe implements Risor's os.File interface.
// The type is backed by Go's io.Pipe therefore it allows concurrent read/write.
type Pipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func NewPipe() *Pipe {
	r, w := io.Pipe()
	return &Pipe{
		r: r,
		w: w,
	}
}

func (f *Pipe) Write(p []byte) (int, error) {
	return f.w.Write(p)
}

func (f *Pipe) Stat() (risoros.FileInfo, error) {
	return risoros.NewFileInfo(risoros.GenericFileInfoOpts{
		Name:    "grr",
		Size:    0,
		Mode:    0,
		ModTime: time.Time{},
		IsDir:   false,
	}), nil
}

func (f *Pipe) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

func (f *Pipe) Close() error {
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
