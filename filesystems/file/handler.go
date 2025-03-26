package file

import (
	"net/url"

	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/foojank/internal/uri"
)

type URIHandler struct {
	fs risoros.FS
}

func NewURIHandler() (*URIHandler, error) {
	fs, err := NewFS()
	if err != nil {
		return nil, err
	}

	return &URIHandler{
		fs: fs,
	}, nil
}

func (h *URIHandler) GetFS(u *url.URL) (risoros.FS, string, error) {
	pth := uri.ToPath(u)
	return h.fs, pth, nil
}
