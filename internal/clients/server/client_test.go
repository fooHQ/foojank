package server_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/jwt/v2"
	natssrv "github.com/nats-io/nats-server/v2/server"
	natstest "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/server"
)

func TestOnConnectReusesConnection(t *testing.T) {
	env := startOperatorServer(t)
	userKP := mustUserKey(t)
	userPub := publicKey(t, userKP)

	limited, limitedSeed := env.userCreds(t, userKP, jwt.Permissions{
		Pub: jwt.Permission{Allow: []string{"issue.jwt"}},
		Sub: jwt.Permission{Allow: []string{auth.InboxPrefix(userPub) + ".>"}},
	})
	upgraded, _ := env.userCreds(t, userKP, jwt.Permissions{
		Pub: jwt.Permission{Allow: []string{"upgraded.subject"}},
		Sub: jwt.Permission{Allow: []string{auth.InboxPrefix(userPub) + ".>"}},
	})
	upgradedClaims, err := jwt.DecodeUserClaims(upgraded)
	require.NoError(t, err)
	upgradedClaims.Expires = time.Now().Add(15 * time.Minute).Unix()
	upgraded, err = upgradedClaims.Encode(env.account)
	require.NoError(t, err)

	daemonNC := env.connect(t, jwt.Permissions{})
	_, err = daemonNC.Subscribe("issue.jwt", func(msg *nats.Msg) {
		_ = msg.Respond([]byte(upgraded))
	})
	require.NoError(t, err)
	received := make(chan []byte, 1)
	_, err = daemonNC.Subscribe("upgraded.subject", func(msg *nats.Msg) {
		received <- msg.Data
	})
	require.NoError(t, err)
	require.NoError(t, daemonNC.Flush())

	var connected *nats.Conn
	var calls int
	client, err := server.New(
		context.Background(),
		[]string{env.srv.ClientURL()},
		limited,
		limitedSeed,
		"",
		server.WithOnConnect(func(_ context.Context, nc *nats.Conn) (string, error) {
			calls++
			connected = nc
			msg, err := nc.Request("issue.jwt", nil, 2*time.Second)
			if err != nil {
				return "", err
			}
			return string(msg.Data), nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { client.Conn().Close() })

	require.Equal(t, 1, calls)
	require.Same(t, connected, client.Conn())
	require.True(t, client.Conn().IsConnected())
	require.Equal(t, userPub, client.UserID())
	require.Eventually(t, func() bool {
		return env.srv.NumClients() == 2
	}, 2*time.Second, 10*time.Millisecond)

	require.NoError(t, client.Conn().Publish("upgraded.subject", []byte("ok")))
	select {
	case data := <-received:
		require.Equal(t, "ok", string(data))
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for republished message")
	}

	errCh := make(chan error, 1)
	client.Conn().SetErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
		select {
		case errCh <- err:
		default:
		}
	})
	require.NoError(t, client.Conn().Publish("issue.jwt", []byte("x")))
	require.NoError(t, client.Conn().FlushTimeout(time.Second))
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, nats.ErrPermissionViolation)
	case <-time.After(2 * time.Second):
		t.Fatal("expected a permission violation for the previous jwt")
	}
}

func TestOnConnectRejectsEmptyJWT(t *testing.T) {
	env := startOperatorServer(t)
	userKP := mustUserKey(t)
	token, seed := env.userCreds(t, userKP, jwt.Permissions{})

	_, err := server.New(
		context.Background(),
		[]string{env.srv.ClientURL()},
		token,
		seed,
		"",
		server.WithOnConnect(func(context.Context, *nats.Conn) (string, error) {
			return "", nil
		}),
	)
	require.ErrorContains(t, err, "invalid user JWT")
}

func TestOnConnectCallbackError(t *testing.T) {
	env := startOperatorServer(t)
	userKP := mustUserKey(t)
	token, seed := env.userCreds(t, userKP, jwt.Permissions{})

	_, err := server.New(
		context.Background(),
		[]string{env.srv.ClientURL()},
		token,
		seed,
		"",
		server.WithOnConnect(func(context.Context, *nats.Conn) (string, error) {
			return "", errors.New("boom")
		}),
	)
	require.EqualError(t, err, "boom")
	require.Eventually(t, func() bool {
		return env.srv.NumClients() == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestOnConnectRejectsInvalidJWT(t *testing.T) {
	env := startOperatorServer(t)
	userKP := mustUserKey(t)
	userPub := publicKey(t, userKP)
	token, seed := env.userCreds(t, userKP, jwt.Permissions{})

	other := mustUserKey(t)
	otherJWT, _ := env.userCreds(t, other, jwt.Permissions{})

	expiredClaims, err := jwt.DecodeUserClaims(token)
	require.NoError(t, err)
	expiredClaims.Expires = time.Now().Add(-time.Minute).Unix()
	expiredJWT, err := expiredClaims.Encode(env.account)
	require.NoError(t, err)

	notBeforeClaims, err := jwt.DecodeUserClaims(token)
	require.NoError(t, err)
	notBeforeClaims.NotBefore = time.Now().Add(time.Hour).Unix()
	notBeforeJWT, err := notBeforeClaims.Encode(env.account)
	require.NoError(t, err)

	tests := []struct {
		name string
		jwt  string
		want string
		auth bool
	}{
		{name: "malformed", jwt: "not-a-jwt", want: "invalid user JWT"},
		{name: "subject", jwt: otherJWT, want: "does not match \"" + userPub + "\""},
		{name: "expired", jwt: expiredJWT, auth: true},
		{name: "not yet valid", jwt: notBeforeJWT, auth: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.New(
				context.Background(),
				[]string{env.srv.ClientURL()},
				token,
				seed,
				"",
				server.WithOnConnect(func(context.Context, *nats.Conn) (string, error) {
					return tt.jwt, nil
				}),
			)
			if tt.auth {
				require.ErrorIs(t, err, nats.ErrAuthorization)
				return
			}
			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestOnConnectAuthFailure(t *testing.T) {
	env := startOperatorServer(t)
	userKP := mustUserKey(t)
	token, seed := env.userCreds(t, userKP, jwt.Permissions{})

	otherAccount, err := nkeys.CreateAccount()
	require.NoError(t, err)
	claims, err := jwt.DecodeUserClaims(token)
	require.NoError(t, err)
	badJWT, err := claims.Encode(otherAccount)
	require.NoError(t, err)

	_, err = server.New(
		context.Background(),
		[]string{env.srv.ClientURL()},
		token,
		seed,
		"",
		server.WithOnConnect(func(context.Context, *nats.Conn) (string, error) {
			return badJWT, nil
		}),
	)
	require.ErrorIs(t, err, nats.ErrAuthorization)
	require.Eventually(t, func() bool {
		return env.srv.NumClients() == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestNewWithCredsFileOnConnect(t *testing.T) {
	env := startOperatorServer(t)
	userKP := mustUserKey(t)
	userPub := publicKey(t, userKP)
	token, seed := env.userCreds(t, userKP, jwt.Permissions{})

	dir := t.TempDir()
	path := filepath.Join(dir, "user.creds")
	require.NoError(t, os.WriteFile(path, decorateCreds(t, token, []byte(seed)), 0o600))

	called := false
	client, err := server.NewWithCredsFile(
		context.Background(),
		[]string{env.srv.ClientURL()},
		path,
		"",
		server.WithOnConnect(func(_ context.Context, nc *nats.Conn) (string, error) {
			called = true
			require.True(t, nc.IsConnected())
			return token, nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { client.Conn().Close() })
	require.True(t, called)
	require.Equal(t, userPub, client.UserID())
}

func TestNewWithoutOnConnect(t *testing.T) {
	env := startOperatorServer(t)
	userKP := mustUserKey(t)
	token, seed := env.userCreds(t, userKP, jwt.Permissions{})

	client, err := server.New(context.Background(), []string{env.srv.ClientURL()}, token, seed, "")
	require.NoError(t, err)
	t.Cleanup(func() { client.Conn().Close() })
	require.True(t, client.Conn().IsConnected())
	require.Equal(t, publicKey(t, userKP), client.UserID())
}

type operatorEnv struct {
	srv     *natssrv.Server
	account nkeys.KeyPair
}

func startOperatorServer(t *testing.T) *operatorEnv {
	t.Helper()

	operatorKP, err := nkeys.CreateOperator()
	require.NoError(t, err)
	operatorPub, err := operatorKP.PublicKey()
	require.NoError(t, err)

	accountKP, err := nkeys.CreateAccount()
	require.NoError(t, err)
	accountPub, err := accountKP.PublicKey()
	require.NoError(t, err)

	accountClaims := jwt.NewAccountClaims(accountPub)
	accountClaims.Name = "test"
	accountJWT, err := accountClaims.Encode(operatorKP)
	require.NoError(t, err)

	opts := natstest.DefaultTestOptions
	opts.Port = -1
	opts.NoLog = true
	opts.TrustedKeys = []string{operatorPub}

	srv := natstest.RunServer(&opts)
	t.Cleanup(srv.Shutdown)

	resolver := &natssrv.MemAccResolver{}
	require.NoError(t, resolver.Store(accountPub, accountJWT))
	srv.SetAccountResolver(resolver)

	return &operatorEnv{
		srv:     srv,
		account: accountKP,
	}
}

func (e *operatorEnv) userCreds(t *testing.T, user nkeys.KeyPair, perms jwt.Permissions) (string, string) {
	t.Helper()

	pub, err := user.PublicKey()
	require.NoError(t, err)

	claims := jwt.NewUserClaims(pub)
	claims.Permissions = perms
	token, err := claims.Encode(e.account)
	require.NoError(t, err)

	seed, err := user.Seed()
	require.NoError(t, err)
	return token, string(seed)
}

func (e *operatorEnv) connect(t *testing.T, perms jwt.Permissions) *nats.Conn {
	t.Helper()

	userKP := mustUserKey(t)
	token, seed := e.userCreds(t, userKP, perms)
	nc, err := nats.Connect(e.srv.ClientURL(), nats.UserJWTAndSeed(token, seed))
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	return nc
}

func mustUserKey(t *testing.T) nkeys.KeyPair {
	t.Helper()
	kp, err := nkeys.CreateUser()
	require.NoError(t, err)
	return kp
}

func publicKey(t *testing.T, kp nkeys.KeyPair) string {
	t.Helper()
	pub, err := kp.PublicKey()
	require.NoError(t, err)
	return pub
}

func decorateCreds(t *testing.T, token string, seed []byte) []byte {
	t.Helper()
	data, err := jwt.DecorateJWT(token)
	require.NoError(t, err)
	decorated, err := jwt.DecorateSeed(seed)
	require.NoError(t, err)
	return append(data, decorated...)
}
