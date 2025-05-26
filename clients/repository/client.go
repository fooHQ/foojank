package repository

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/repository"
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
		return &Error{err}
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, repository string) error {
	err := c.js.DeleteObjectStore(ctx, repository)
	if err != nil {
		return &Error{err}
	}
	return nil
}

func (c *Client) List(ctx context.Context) ([]*Repository, error) {
	var result []*Repository
	for r := range c.js.ObjectStores(ctx).Status() {
		result = append(result, &Repository{
			Name:        r.Bucket(),
			Description: r.Description(),
			Size:        r.Size(),
		})
	}
	return result, nil
}

func (c *Client) Get(ctx context.Context, name string) (*repository.Repository, error) {
	s, err := c.js.ObjectStore(ctx, name)
	if err != nil {
		return nil, &Error{err}
	}

	r, err := repository.New(ctx, s)
	if err != nil {
		return nil, &Error{err}
	}

	return r, nil
}

func (c *Client) PutFile(ctx context.Context, repository, filename string, reader io.Reader) error {
	s, err := c.js.ObjectStore(ctx, repository)
	if err != nil {
		return &Error{err}
	}

	_, err = s.Put(ctx, jetstream.ObjectMeta{
		Name: filename,
	}, reader)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetFile(ctx context.Context, repository, filename string) (*File, error) {
	s, err := c.js.ObjectStore(ctx, repository)
	if err != nil {
		return nil, &Error{err}
	}

	res, err := s.Get(ctx, filename)
	if err != nil {
		return nil, &Error{err}
	}
	defer res.Close()

	b, err := io.ReadAll(res)
	if err != nil {
		return nil, err
	}

	info, err := res.Info()
	if err != nil {
		return nil, err
	}

	return &File{
		b:        bytes.NewReader(b),
		Name:     info.Name,
		Size:     info.Size,
		Modified: info.ModTime,
	}, nil
}

func (c *Client) DeleteFile(ctx context.Context, repository, filename string) error {
	s, err := c.js.ObjectStore(ctx, repository)
	if err != nil {
		return &Error{err}
	}

	err = s.Delete(ctx, filename)
	if err != nil {
		return &Error{err}
	}

	return nil
}

func (c *Client) ListFiles(ctx context.Context, repository string) ([]*File, error) {
	s, err := c.js.ObjectStore(ctx, repository)
	if err != nil {
		return nil, &Error{err}
	}

	files, err := s.List(ctx)
	if err != nil {
		if errors.Is(err, jetstream.ErrNoObjectsFound) {
			return nil, nil
		}
		return nil, &Error{err}
	}

	var result []*File
	for i := range files {
		if files[i].Deleted {
			continue
		}

		result = append(result, &File{
			Name:     files[i].Name,
			Size:     files[i].Size,
			Modified: files[i].ModTime,
		})
	}

	return result, nil
}

type Error struct {
	err error
}

func (e *Error) Error() string {
	switch {
	case errors.Is(e.err, jetstream.ErrBucketExists):
		return "repository already exists"
	case errors.Is(e.err, jetstream.ErrBucketNotFound):
		return "repository not found"
	case errors.Is(e.err, jetstream.ErrObjectNotFound):
		return "file not found"
	case errors.Is(e.err, jetstream.ErrInvalidStoreName):
		return "invalid repository name"
	}
	return e.err.Error()
}
