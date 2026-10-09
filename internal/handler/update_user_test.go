package handler_test

import (
	"context"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/handler"
	"github.com/foohq/foojank/internal/log"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type testMsg struct {
	data any
}

func (m testMsg) ID() string           { return "" }
func (m testMsg) Subject() string      { return "" }
func (m testMsg) ReplySubject() string { return "" }
func (m testMsg) Data() any            { return m.data }
func (m testMsg) Ack() error           { return nil }

func newTestJetStream(t *testing.T) jetstream.JetStream {
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

func TestUpdateUser(t *testing.T) {
	js := newTestJetStream(t)
	ctx := context.Background()

	dir, err := directory.OpenUserDirectory(ctx, js)
	require.NoError(t, err)

	created := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	_, err = dir.Create(ctx, directory.UserDirectoryEntry{
		ID:          "id-alice",
		Name:        "alice",
		Description: "first",
		Privileges:  []string{"USER.CREATE", "USER.GET"},
		CreatedAt:   created,
	})
	require.NoError(t, err)

	h := handler.NewNATSHandler(log.NewLogger("error", true), handler.NATSHandlerConfig{
		UserDirectory: dir,
	})

	update := func(req any) protodaemon.UpdateUserResponse {
		t.Helper()
		resp, ok := h.UpdateUser(ctx, nil, testMsg{data: req}).(protodaemon.UpdateUserResponse)
		require.True(t, ok)
		return resp
	}

	resp := update("nope")
	require.EqualError(t, resp.Error, "invalid request data")

	resp = update(protodaemon.UpdateUserRequest{})
	require.EqualError(t, resp.Error, "name is required")

	resp = update(protodaemon.UpdateUserRequest{Name: "missing"})
	require.EqualError(t, resp.Error, `"missing" not found`)

	resp = update(protodaemon.UpdateUserRequest{
		Name:          "alice",
		Description:   "ignored",
		IsDescription: false,
		SetPrivileges: []string{"USER.LIST"},
		UnsetPrivileges: []string{
			"user.create",
		},
	})
	require.NoError(t, resp.Error)

	got, err := dir.Get(ctx, "alice")
	require.NoError(t, err)
	require.Equal(t, "id-alice", got.ID)
	require.Equal(t, "alice", got.Name)
	require.Equal(t, "first", got.Description)
	require.Equal(t, []string{"USER.GET", "USER.LIST"}, got.Privileges)
	require.True(t, created.Equal(got.CreatedAt))

	resp = update(protodaemon.UpdateUserRequest{
		Name:          "alice",
		Description:   "",
		IsDescription: true,
	})
	require.NoError(t, resp.Error)

	got, err = dir.Get(ctx, "alice")
	require.NoError(t, err)
	require.Equal(t, "id-alice", got.ID)
	require.Equal(t, "alice", got.Name)
	require.Empty(t, got.Description)
	require.Equal(t, []string{"USER.GET", "USER.LIST"}, got.Privileges)

	resp = update(protodaemon.UpdateUserRequest{
		Name:          "alice",
		SetPrivileges: []string{"NOPE"},
	})
	require.Error(t, resp.Error)
	got, err = dir.Get(ctx, "alice")
	require.NoError(t, err)
	require.Equal(t, []string{"USER.GET", "USER.LIST"}, got.Privileges)
}
