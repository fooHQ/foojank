package filefs

import (
	"net/url"

	risoros "github.com/risor-io/risor/os"

	engineos "github.com/foohq/foojank/internal/engine/os"
)

var _ engineos.URIHandler = &URIHandler{}

type URIHandler struct {
	fs risoros.FS
}

func NewURIHandler() *URIHandler {
	return &URIHandler{
		fs: NewFS(),
	}
}

func (h *URIHandler) GetFS(u *url.URL) (risoros.FS, error) {
	return h.fs, nil
}
