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
	"github.com/nats-io/nats.go/micro"
	"github.com/stretchr/testify/require"
)

var _ micro.Request = &Request{}

type Request struct {
	FSubject      string
	FReplySubject string
	FData         []byte
	ResponseCh    chan []byte
}

func (r Request) Respond(bytes []byte, opt ...micro.RespondOpt) error {
	r.ResponseCh <- bytes
	return nil
}

func (r Request) RespondJSON(a any, opt ...micro.RespondOpt) error {
	panic("implement me")
}

func (r Request) Error(code, description string, data []byte, opts ...micro.RespondOpt) error {
	s := fmt.Sprintf("%s: %s", code, description)
	r.ResponseCh <- []byte(s)
	return nil
}

func (r Request) Data() []byte {
	return r.FData
}

func (r Request) Headers() micro.Headers {
	panic("implement me")
}

func (r Request) Subject() string {
	return r.FSubject
}

func (r Request) Reply() string {
	return r.FReplySubject
}

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

func NewNatsObjectStore(t *testing.T, nc *nats.Conn) jetstream.ObjectStore {
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	s, err := js.CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
		Bucket: fmt.Sprintf("test_bucket_%d", rand.Int()),
	})
	require.NoError(t, err)
	return s
}
