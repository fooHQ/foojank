package nats

import (
	"net/url"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/foojank/internal/uri"
)

type URIHandler struct {
	fs risoros.FS
}

func NewURIHandler(store jetstream.ObjectStore) *URIHandler {
	return &URIHandler{
		fs: NewFS(store),
	}
}

func (h *URIHandler) GetFS(u *url.URL) (risoros.FS, string, error) {
	pth := uri.ToPath(u)
	return h.fs, pth, nil
}
