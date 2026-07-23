package server

import (
	"crypto/x509"
	"os"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	userID string
	jetstream.JetStream
}

func New(servers []string, userJWT, userSeed, serverCert string) (*Client, error) {
	js, err := connect(strings.Join(servers, ","), userJWT, userSeed, serverCert)
	if err != nil {
		return nil, err
	}

	claims, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return nil, err
	}

	return &Client{
		JetStream: js,
		userID:    claims.Subject,
	}, nil
}

func (c *Client) UserID() string {
	return c.userID
}

func connect(server, userJWT, userKey, serverCert string) (jetstream.JetStream, error) {
	opts := []nats.Option{
		nats.MaxReconnects(-1),
	}

	if userJWT != "" && userKey != "" {
		opts = append(opts, nats.UserJWTAndSeed(userJWT, userKey))
	}

	if serverCert != "" {
		opts = append(opts, nats.TLSHandshakeFirst())
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
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(b)
		return pool, nil
	}
}
