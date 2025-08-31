package server

import (
	"context"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/sstls"
)

type Client struct {
	js jetstream.JetStream
}

func New(
	servers []string,
	userJWT,
	userKey,
	caCertFile string,
) (*Client, error) {
	nc, err := nats.Connect(
		strings.Join(servers, ","),
		nats.UserJWTAndSeed(userJWT, userKey),
		nats.MaxReconnects(-1),
		nats.ClientTLSConfig(nil, sstls.DecodeCertificateHandler(caCertFile)),
		/*nats.ConnectHandler(func(_ *nats.Conn) {
			logger.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			logger.Info("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			err = fmt.Errorf("disconnected from the server: %w", err)
			logger.Warn(err.Error())
		}),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			err = fmt.Errorf("server error: %w", err)
			logger.Warn(err.Error())
		}),*/
	)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return &Client{
		js: js,
	}, nil
}

func (c *Client) Publish(ctx context.Context, msg *nats.Msg) error {
	if msg.Header == nil {
		msg.Header = make(nats.Header)
	}
	msg.Header.Set(nats.MsgIdHdr, nuid.Next())
	_, err := c.js.PublishMsg(ctx, msg)
	return err
}

func (c *Client) CreateStream(ctx context.Context, name string, subjects []string) (jetstream.Stream, error) {
	_, err := c.js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     name,
		Subjects: subjects,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *Client) ListStreams(ctx context.Context) ([]string, error) {
	var names []string
	for stream := range c.js.ListStreams(ctx).Info() {
		if stream == nil {
			break
		}
		names = append(names, stream.Config.Name)
	}
	return names, nil
}

func (c *Client) CreateConsumer(ctx context.Context, stream string) (jetstream.Consumer, error) {
	consumer, err := c.js.CreateConsumer(ctx, stream, jetstream.ConsumerConfig{
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
	})
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func (c *Client) CreateDurableConsumer(ctx context.Context, stream, name string, subjects []string) (jetstream.Consumer, error) {
	consumer, err := c.js.CreateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Durable:        name,
		DeliverPolicy:  jetstream.DeliverLastPolicy,
		AckPolicy:      jetstream.AckExplicitPolicy,
		MaxAckPending:  1,
		FilterSubjects: subjects,
	})
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func (c *Client) CreateObjectStore(ctx context.Context, name, description string) error {
	_, err := c.js.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{
		Bucket:      name,
		Description: description,
	})
	if err != nil {
		return &errorObjectStore{err}
	}
	return nil
}

func (c *Client) DeleteObjectStore(ctx context.Context, name string) error {
	err := c.js.DeleteObjectStore(ctx, name)
	if err != nil {
		return &errorObjectStore{err}
	}
	return nil
}

func (c *Client) ListObjectStores(ctx context.Context) ([]*ObjectStore, error) {
	var result []*ObjectStore
	for name := range c.js.ObjectStoreNames(ctx).Name() {
		store, err := c.js.ObjectStore(ctx, name)
		if err != nil {
			return nil, &errorObjectStore{err}
		}

		s, err := NewObjectStore(ctx, store)
		if err != nil {
			return nil, err
		}

		result = append(result, s)
	}
	return result, nil
}

func (c *Client) GetObjectStore(ctx context.Context, name string) (*ObjectStore, error) {
	store, err := c.js.ObjectStore(ctx, name)
	if err != nil {
		return nil, &errorObjectStore{err}
	}

	s, err := NewObjectStore(ctx, store)
	if err != nil {
		return nil, err
	}

	return s, nil
}
