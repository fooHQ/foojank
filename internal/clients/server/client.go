package server

import (
	"crypto/x509"
	"os"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/auth"
)

type Client struct {
	jetstream.JetStream
	userID string
}

func New(servers []string, userJWT, userSeed, serverCert string) (*Client, error) {
	userClaims, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return nil, err
	}

	inboxPrefix := auth.InboxPrefix(userClaims.Subject)

	opts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.CustomInboxPrefix(inboxPrefix),
	}

	if userJWT != "" && userSeed != "" {
		opts = append(opts, nats.UserJWTAndSeed(userJWT, userSeed))
	}

	if serverCert != "" {
		opts = append(opts, nats.TLSHandshakeFirst())
		b, err := os.ReadFile(serverCert)
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.ClientTLSConfig(nil, decodeCertificatesHandler(b)))
	}

	js, err := connect(servers, opts)
	if err != nil {
		return nil, err
	}

	return &Client{
		JetStream: js,
		userID:    userClaims.Subject,
	}, nil
}

func NewWithCredsFile(servers []string, credsFile string, serverCert string) (*Client, error) {
	b, err := os.ReadFile(credsFile)
	if err != nil {
		return nil, err
	}

	userJWT, err := jwt.ParseDecoratedJWT(b)
	if err != nil {
		return nil, err
	}

	userKey, err := jwt.ParseDecoratedUserNKey(b)
	if err != nil {
		return nil, err
	}

	userSeed, err := userKey.Seed()
	if err != nil {
		return nil, err
	}

	return New(servers, userJWT, string(userSeed), serverCert)
}

func (c *Client) UserID() string {
	return c.userID
}

func connect(servers []string, opts []nats.Option) (jetstream.JetStream, error) {
	nc, err := nats.Connect(strings.Join(servers, ","), opts...)
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
