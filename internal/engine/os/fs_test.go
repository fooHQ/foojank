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

	filename := "/create/file"
	_, err = fs.Create(filename)
	require.NoError(t, err)

	_, err = fs.Create(filename)
	require.NoError(t, err)
}

func TestFS_Open(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	fs, err := engineos.NewVirtualFS(js, "test")
	require.NoError(t, err)

	filename := fmt.Sprintf("/open/file_%d", rand.Int())
	_, err = fs.Create(filename)
	require.NoError(t, err)

	_, err = fs.Open(filename)
	require.NoError(t, err)

	_, err = fs.Open("/open/notexists")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func Test_ReadFile(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	fs, err := engineos.NewVirtualFS(js, "test")
	require.NoError(t, err)

	filename := fmt.Sprintf("/read_file/file_%d", rand.Int())
	f, err := fs.Create(filename)
	require.NoError(t, err)
	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()

	message := []byte("hello world")
	_, err = f.Write(message)
	require.NoError(t, err)

	b, err := fs.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, message, b)

	_, err = fs.ReadFile("/read_file/notexists")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func Test_Remove(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	fs, err := engineos.NewVirtualFS(js, "test")
	require.NoError(t, err)

	filename := fmt.Sprintf("/remove/file_%d", rand.Int())
	f, err := fs.Create(filename)
	require.NoError(t, err)
	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()

	err = fs.Remove(filename)
	require.NoError(t, err)

	err = fs.Remove("/remove/notexists")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func Test_Rename(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	fs, err := engineos.NewVirtualFS(js, "test")
	require.NoError(t, err)

	filename := fmt.Sprintf("/rename/file_%d", rand.Int())
	f, err := fs.Create(filename)
	require.NoError(t, err)
	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()

	newFilename := filename + "_2"
	err = fs.Rename(filename, newFilename)
	require.NoError(t, err)

	_, err = fs.Open(newFilename)
	require.NoError(t, err)
}

func Test_WriteFile(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	fs, err := engineos.NewVirtualFS(js, "test")
	require.NoError(t, err)

	filename := fmt.Sprintf("/write_file/file_%d", rand.Int())
	message := []byte("hello world")
	err = fs.WriteFile(filename, message, 0)
	require.NoError(t, err)

	b, err := fs.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, message, b)
}
