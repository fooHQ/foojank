package repository

import (
	"context"
	"github.com/nats-io/nats.go/jetstream"
	"io"
)

type Client struct {
	js jetstream.JetStream
}

func New(js jetstream.JetStream) *Client {
	return &Client{
		js: js,
	}
}

func (c *Client) Create(ctx context.Context, name, description string) error {
	_, err := c.js.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{
		Bucket:      name,
		Description: description,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) List(ctx context.Context) ([]Repository, error) {
	listener := c.js.ObjectStores(ctx)
	var result []Repository
loop:
	for {
		select {
		case status, ok := <-listener.Status():
			if !ok {
				break loop
			}

			result = append(result, Repository{
				Name:        status.Bucket(),
				Description: status.Description(),
				Size:        status.Size(),
			})

		case <-ctx.Done():
			break loop
		}
	}
	return result, nil
}

func (c *Client) Push(ctx context.Context, repository, filename string, reader io.Reader) error {
	s, err := c.js.ObjectStore(ctx, repository)
	if err != nil {
		return err
	}

	_, err = s.Put(ctx, jetstream.ObjectMeta{
		Name: filename,
	}, reader)
	if err != nil {
		return err
	}

	return nil
}

/*func (c *Client) Get(ctx context.Context, name string) error {
	s, err := c.js.ObjectStore(ctx, name)
	if err != nil {
		return err
	}

	c.js.ObjectStores()

	o.GetInfo()
}*/
