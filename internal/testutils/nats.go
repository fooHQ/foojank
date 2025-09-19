package testutils

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func NewNatsServer() *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.NoLog = false
	opts.Port = -1 // Pick a random port
	opts.Debug = true
	opts.JetStream = true
	opts.StoreDir = "/tmp/nats-server"
	srv := natsserver.RunServer(&opts)
	return srv
}

func NewNatsServerAndConnection(t *testing.T) (*server.Server, *nats.Conn) {
	s := NewNatsServer()
	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(func() {
		nc.Close()
		s.Shutdown()
	})
	return s, nc
}

func NewJetStreamConnection(t *testing.T) (*server.Server, jetstream.JetStream) {
	s := NewNatsServer()
	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	t.Cleanup(func() {
		nc.Close()
		s.Shutdown()
	})

	return s, js
}

func NewNatsObjectStore(t *testing.T, nc *nats.Conn) jetstream.ObjectStore {
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	s, err := js.CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
		Bucket: fmt.Sprintf("test_bucket_%d", rand.Int()),
	})
	require.NoError(t, err)
	return s
}
