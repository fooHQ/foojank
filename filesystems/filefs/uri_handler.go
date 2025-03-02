package filefs

import (
	"net/url"

	risoros "github.com/risor-io/risor/os"
)

type URIHandler struct {
	fs risoros.FS
}

func NewURIHandler() *URIHandler {
	return &URIHandler{
		fs: NewFS(),
	}
}

func (h *URIHandler) GetFS(u *url.URL) (risoros.FS, string, error) {
	pth := ToPath(u)
	return h.fs, pth, nil
}
