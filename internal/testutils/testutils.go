package testutils

import (
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewNatsServer(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.NoLog = false
	opts.Port = port
	return natsserver.RunServer(&opts)
}

func NewNatsServerAndConnection(t *testing.T) (*server.Server, *nats.Conn) {
	s := NewNatsServer(14444)
	nc, err := nats.Connect(s.ClientURL())
	assert.NoError(t, err)
	t.Cleanup(func() {
		nc.Close()
		s.Shutdown()
	})
	return s, nc
}
