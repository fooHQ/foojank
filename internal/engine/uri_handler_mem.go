package engine

import (
	"net/url"

	risoros "github.com/risor-io/risor/os"

	memfs "github.com/foohq/foojank/filesystems/mem"
	"github.com/foohq/foojank/internal/uri"
)

const URIMem = "mem"

type MemURIHandler struct {
	fs *memfs.FS
}

func NewMemURIHandler() (*MemURIHandler, error) {
	fs, err := memfs.NewFS()
	if err != nil {
		return nil, err
	}

	return &MemURIHandler{
		fs: fs,
	}, nil
}

func (h *MemURIHandler) GetFS(u *url.URL) (risoros.FS, string, error) {
	pth := uri.ToPath(u)
	return h.fs, pth, nil
}

func (h *MemURIHandler) Close() error {
	return nil
}
