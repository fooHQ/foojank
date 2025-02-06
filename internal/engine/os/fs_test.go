package os_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	engineos "github.com/foohq/foojank/internal/engine/os"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
)

func TestFS_Create(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	fs, err := engineos.NewVirtualFS(js, "test")
	require.NoError(t, err)

	_, err = fs.Create("/create/file")
	require.NoError(t, err)
}

func TestFS_Open(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	fs, err := engineos.NewVirtualFS(js, "test")
	require.NoError(t, err)

	filename := fmt.Sprintf("/open/file_%d", rand.Int())
	_, err = fs.Open(filename)
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = fs.Create(filename)
	require.NoError(t, err)

	f, err := fs.Open(filename)
	require.NoError(t, err)
	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()

	const textContent = "hello world"
	n, err := f.Write([]byte(textContent))
	require.NoError(t, err)
	require.NotZero(t, n)

	b := make([]byte, len(textContent))
	n, err = f.Read(b)
	require.NoError(t, err)
	require.NotZero(t, n)
	require.Equal(t, textContent, string(b))
}
