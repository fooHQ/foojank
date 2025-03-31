package engine

import (
	"net/url"

	risoros "github.com/risor-io/risor/os"

	filefs "github.com/foohq/foojank/filesystems/file"
	"github.com/foohq/foojank/internal/uri"
)

const URIFile = "mem"

type FileURIHandler struct {
	fs *filefs.FS
}

func NewFileURIHandler() (*FileURIHandler, error) {
	fs, err := filefs.NewFS()
	if err != nil {
		return nil, err
	}

	return &FileURIHandler{
		fs: fs,
	}, nil
}

func (h *FileURIHandler) GetFS(u *url.URL) (risoros.FS, string, error) {
	pth := uri.ToPath(u)
	return h.fs, pth, nil
}

func (h *FileURIHandler) Close() error {
	return nil
}
