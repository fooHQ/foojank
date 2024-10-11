package natsfs

import (
	"github.com/foojank/foojank/internal/testutils"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFS_Create(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	assert.NoError(t, err)
	fs, err := NewFS(js, "test")
	assert.NoError(t, err)
	_, err = fs.Create("/create/file")
	assert.NoError(t, err)
}

func TestFS_Open(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	assert.NoError(t, err)
	fs, err := NewFS(js, "test")
	assert.NoError(t, err)
	_, err = fs.Open("/open/file")
	assert.NoError(t, err)
	_, err = fs.Create("/open/file")
	assert.NoError(t, err)
	_, err = fs.Open("/open/file")
	assert.NoError(t, err)
}
