package directory

import (
	"context"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func newJetStream(t *testing.T) jetstream.JetStream {
	t.Helper()

	opts := natsserver.DefaultTestOptions
	opts.Port = -1
	opts.JetStream = true
	opts.StoreDir = t.TempDir()
	s := natsserver.RunServer(&opts)
	t.Cleanup(s.Shutdown)

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(nc.Close)

	js, err := jetstream.New(nc)
	require.NoError(t, err)
	return js
}

func TestOpenAgentDirectoryIdempotent(t *testing.T) {
	js := newJetStream(t)
	ctx := context.Background()

	dir, err := OpenAgentDirectory(ctx, js)
	require.NoError(t, err)
	require.NotNil(t, dir.store)

	dir, err = OpenAgentDirectory(ctx, js)
	require.NoError(t, err)
	require.NotNil(t, dir.store)
}

func TestOpenAgentDirectoryBucketExists(t *testing.T) {
	js := newJetStream(t)
	ctx := context.Background()

	_, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      agentDirectoryName,
		Description: "different config",
	})
	require.NoError(t, err)

	_, err = OpenAgentDirectory(ctx, js)
	require.ErrorIs(t, err, ErrStorageExists)
}

func TestDirectoryCreateGetDelete(t *testing.T) {
	js := newJetStream(t)
	ctx := context.Background()

	dir, err := OpenAgentDirectory(ctx, js)
	require.NoError(t, err)

	entry, err := dir.Create(ctx, AgentDirectoryEntry{
		ID:        "agent1",
		Name:      "alpha",
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	require.NotZero(t, entry.Revision)

	got, err := dir.Get(ctx, "agent1")
	require.NoError(t, err)
	require.Equal(t, "agent1", got.ID)
	require.Equal(t, "alpha", got.Name)

	got, err = dir.Get(ctx, "alpha")
	require.NoError(t, err)
	require.Equal(t, "agent1", got.ID)

	_, err = dir.Create(ctx, AgentDirectoryEntry{
		ID:   "agent1",
		Name: "beta",
	})
	require.ErrorIs(t, err, ErrKeyExists)

	_, err = dir.Get(ctx, "missing")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = dir.Create(ctx, AgentDirectoryEntry{
		ID:   "bad key",
		Name: "name",
	})
	require.ErrorIs(t, err, ErrNameInvalid)
}
