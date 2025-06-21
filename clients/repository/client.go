package repository

import (
	"context"
	"errors"

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

func (c *Client) List(ctx context.Context) ([]*repository.Repository, error) {
	var result []*repository.Repository
	for name := range c.js.ObjectStoreNames(ctx).Name() {
		store, err := c.js.ObjectStore(ctx, name)
		if err != nil {
			return nil, &Error{err}
		}

		repo, err := repository.New(ctx, store)
		if err != nil {
			return nil, err
		}

		result = append(result, repo)
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
