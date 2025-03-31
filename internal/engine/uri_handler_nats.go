package engine

import (
	"context"
	"net/url"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"

	natsfs "github.com/foohq/foojank/filesystems/nats"
	"github.com/foohq/foojank/internal/uri"
)

const URINats = "nats"

type NatsURIHandler struct {
	fs *natsfs.FS
}

func NewNatsURIHandler(ctx context.Context, store jetstream.ObjectStore) (*NatsURIHandler, error) {
	fs, err := natsfs.NewFS(ctx, store)
	if err != nil {
		return nil, err
	}

	return &NatsURIHandler{
		fs: fs,
	}, nil
}

func (h *NatsURIHandler) GetFS(u *url.URL) (risoros.FS, string, error) {
	pth := uri.ToPath(u)
	return h.fs, pth, nil
}

func (h *NatsURIHandler) Close() error {
	return h.fs.Close()
}
