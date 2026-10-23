package volume

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/clients/server"
)

type Client struct {
	srv *server.Client
}

func New(srv *server.Client) *Client {
	return &Client{
		srv: srv,
	}
}

func (c *Client) Create(ctx context.Context, name, description string) error {
	_, err := c.srv.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{
		Bucket:      name,
		Description: description,
	})
	if err != nil {
		return translate(err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, name string) error {
	err := c.srv.DeleteObjectStore(ctx, name)
	if err != nil {
		return translate(err)
	}
	return nil
}

func (c *Client) List(ctx context.Context) ([]*Volume, error) {
	var result []*Volume
	for name := range c.srv.ObjectStoreNames(ctx).Name() {
		s, err := c.Get(ctx, name)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (c *Client) Get(ctx context.Context, name string) (*Volume, error) {
	store, err := c.srv.ObjectStore(ctx, name)
	if err != nil {
		return nil, translate(err)
	}

	s, err := NewVolume(ctx, name, store)
	if err != nil {
		return nil, err
	}

	return s, nil
}
