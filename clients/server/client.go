package server

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"
)

type Client struct {
	js jetstream.JetStream
}

func New(
	servers []string,
	userJWT,
	userSeed,
	serverCert string,
) (*Client, error) {
	js, err := connect(strings.Join(servers, ","), userJWT, userSeed, serverCert)
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

func (c *Client) CreateConsumer(ctx context.Context, stream string, subjects []string) (jetstream.Consumer, error) {
	consumer, err := c.js.CreateConsumer(ctx, stream, jetstream.ConsumerConfig{
		DeliverPolicy:  jetstream.DeliverAllPolicy,
		AckPolicy:      jetstream.AckExplicitPolicy,
		MaxAckPending:  1,
		FilterSubjects: subjects,
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

func connect(
	server,
	userJWT,
	userKey,
	serverCert string,
) (jetstream.JetStream, error) {
	opts := []nats.Option{
		nats.MaxReconnects(-1),
	}

	if userJWT != "" && userKey != "" {
		opts = append(opts, nats.UserJWTAndSeed(userJWT, userKey))
	}

	if serverCert != "" {
		b, err := os.ReadFile(serverCert)
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.ClientTLSConfig(nil, decodeCertificatesHandler(b)))
	}

	nc, err := nats.Connect(server, opts...)
	if err != nil {
		return nil, err
	}

	jetStream, err := jetstream.New(
		nc,
		jetstream.WithDefaultTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}

	return jetStream, nil
}

func decodeCertificatesHandler(b []byte) func() (*x509.CertPool, error) {
	return func() (*x509.CertPool, error) {
		var certs []*x509.Certificate
		for len(b) > 0 {
			block, rest := pem.Decode(b)
			if block == nil {
				break
			}
			b = rest

			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, err
				}
				certs = append(certs, cert)
			}
		}

		if len(certs) == 0 {
			return nil, errors.New("no certificates found")
		}

		pool := x509.NewCertPool()
		for _, cert := range certs {
			pool.AddCert(cert)
		}

		return pool, nil
	}
}
