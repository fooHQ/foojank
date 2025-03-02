package natsfs

import (
	"net/url"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"

	engineos "github.com/foohq/foojank/internal/engine/os"
)

var _ engineos.URIHandler = &URIHandler{}

type URIHandler struct {
	fs risoros.FS
}

func NewURIHandler(store jetstream.ObjectStore) *URIHandler {
	return &URIHandler{
		fs: NewFS(store),
	}
}

func (h *URIHandler) GetFS(u *url.URL) (risoros.FS, error) {
	return h.fs, nil
}
