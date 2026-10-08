package server

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"

	"github.com/foohq/foojank/internal/auth"
)

// OnConnectFunc is called once the client is connected with the credentials
// given to New.
//
// A returned JWT replaces the current user JWT. The same *nats.Conn is
// reauthenticated with it; the nkey does not change. An empty JWT keeps the
// current credentials.
//
// The function runs only after the initial connection. Later reconnects keep
// using the JWT already installed.
type OnConnectFunc func(ctx context.Context, nc *nats.Conn) (string, error)

// Option configures a client.
type Option func(*clientOptions)

type clientOptions struct {
	onConnect OnConnectFunc
}

// WithOnConnect sets the function called after the initial connection succeeds.
func WithOnConnect(fn OnConnectFunc) Option {
	return func(o *clientOptions) {
		o.onConnect = fn
	}
}

type Client struct {
	jetstream.JetStream
	userID string
}

func New(ctx context.Context, servers []string, userJWT, userSeed, serverCert string, opts ...Option) (*Client, error) {
	var o clientOptions
	for _, opt := range opts {
		opt(&o)
	}

	userClaims, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return nil, err
	}

	userID := userClaims.Subject

	creds := &credentials{
		token: userJWT,
		seed:  userSeed,
	}

	natsOpts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.CustomInboxPrefix(auth.InboxPrefix(userID)),
	}
	if userJWT != "" && userSeed != "" {
		natsOpts = append(natsOpts, nats.UserJWT(creds.userJWT, creds.sign))
	}

	if serverCert != "" {
		natsOpts = append(natsOpts, nats.TLSHandshakeFirst())
		b, err := os.ReadFile(serverCert)
		if err != nil {
			return nil, err
		}
		natsOpts = append(natsOpts, nats.ClientTLSConfig(nil, decodeCertificatesHandler(b)))
	}

	nc, err := nats.Connect(strings.Join(servers, ","), natsOpts...)
	if err != nil {
		return nil, err
	}

	if o.onConnect != nil {
		newJWT, err := o.onConnect(ctx, nc)
		if err != nil {
			nc.Close()
			return nil, err
		}

		err = validateUserJWT(newJWT, userID)
		if err != nil {
			return nil, err
		}

		creds.setToken(newJWT)

		err = reauthenticate(ctx, nc)
		if err != nil {
			return nil, err
		}
	}

	js, err := jetstream.New(nc, jetstream.WithDefaultTimeout(10*time.Second))
	if err != nil {
		nc.Close()
		return nil, err
	}

	return &Client{
		JetStream: js,
		userID:    userID,
	}, nil
}

func NewWithCredsFile(ctx context.Context, servers []string, credsFile string, serverCert string, opts ...Option) (*Client, error) {
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

	return New(ctx, servers, userJWT, string(userSeed), serverCert, opts...)
}

func (c *Client) UserID() string {
	return c.userID
}

func validateUserJWT(userJWT, userID string) error {
	claims, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return fmt.Errorf("invalid user JWT: %w", err)
	}

	if claims.Subject != userID {
		return fmt.Errorf("user JWT subject %q does not match %q", claims.Subject, userID)
	}

	return nil
}

// reauthenticate forces the open connection to connect again. The user JWT
// callback installed on the connection supplies the current token, so the
// server sees the replacement credentials without a new *nats.Conn.
func reauthenticate(ctx context.Context, nc *nats.Conn) error {
	statuses := nc.StatusChanged(nats.CONNECTED, nats.CLOSED)
	defer nc.RemoveStatusListener(statuses)

	errCh := make(chan error, 1)
	prev := nc.ErrorHandler()
	nc.SetErrorHandler(func(conn *nats.Conn, sub *nats.Subscription, err error) {
		if prev != nil {
			prev(conn, sub, err)
		}
		if !isAuthError(err) {
			return
		}
		select {
		case errCh <- err:
		default:
		}
	})
	defer nc.SetErrorHandler(prev)

	err := nc.ForceReconnect()
	if err != nil {
		return err
	}

	select {
	case status := <-statuses:
		switch status {
		case nats.CONNECTED:
			return nil
		case nats.CLOSED:
			err := nc.LastError()
			if err != nil {
				return err
			}
			return nats.ErrConnectionClosed
		default:
			return fmt.Errorf("unexpected connection status %s", status)
		}

	case <-ctx.Done():
		nc.Close()
		return ctx.Err()

	case err := <-errCh:
		nc.Close()
		return err
	}
}

func isAuthError(err error) bool {
	return errors.Is(err, nats.ErrAuthorization) ||
		errors.Is(err, nats.ErrAuthExpired) ||
		errors.Is(err, nats.ErrAuthRevoked) ||
		errors.Is(err, nats.ErrAccountAuthExpired)
}

type credentials struct {
	mu    sync.Mutex
	token string
	seed  string
}

func (c *credentials) userJWT() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == "" {
		return "", errors.New("user JWT is not set")
	}
	return c.token, nil
}

func (c *credentials) setToken(token string) {
	c.mu.Lock()
	c.token = token
	c.mu.Unlock()
}

func (c *credentials) sign(nonce []byte) ([]byte, error) {
	c.mu.Lock()
	seed := c.seed
	c.mu.Unlock()

	kp, err := nkeys.FromSeed([]byte(seed))
	if err != nil {
		return nil, fmt.Errorf("cannot decode user seed: %w", err)
	}
	defer kp.Wipe()

	return kp.Sign(nonce)
}

func decodeCertificatesHandler(b []byte) func() (*x509.CertPool, error) {
	return func() (*x509.CertPool, error) {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(b)
		return pool, nil
	}
}
