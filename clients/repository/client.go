package repository

import (
	"context"
	"errors"
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

func (c *Client) Delete(ctx context.Context, repository string) error {
	err := c.js.DeleteObjectStore(ctx, repository)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) List(ctx context.Context) ([]Repository, error) {
	var result []Repository
	for status := range c.js.ObjectStores(ctx).Status() {
		result = append(result, Repository{
			Name:        status.Bucket(),
			Description: status.Description(),
			Size:        status.Size(),
		})
	}
	return result, nil
}

func (c *Client) PutFile(ctx context.Context, repository, filename string, reader io.Reader) error {
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

func (c *Client) ListFiles(ctx context.Context, repository string) ([]File, error) {
	s, err := c.js.ObjectStore(ctx, repository)
	if err != nil {
		return nil, err
	}

	files, err := s.List(ctx)
	if err != nil {
		if errors.Is(err, jetstream.ErrNoObjectsFound) {
			return nil, nil
		}
		return nil, err
	}

	var result []File
	for i := range files {
		if files[i].Deleted {
			continue
		}

		result = append(result, File{
			Name:     files[i].Name,
			Size:     files[i].Size,
			Modified: files[i].ModTime,
		})
	}

	return result, nil
}

/*func (c *Client) Get(ctx context.Context, name string) error {
	s, err := c.js.ObjectStore(ctx, name)
	if err != nil {
		return err
	}

	c.js.ObjectStores()

	o.GetInfo()
}*/
