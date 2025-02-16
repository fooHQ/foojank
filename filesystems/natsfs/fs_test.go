package natsfs_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	engineos "github.com/foohq/foojank/filesystems/natsfs"
	"github.com/foohq/foojank/internal/testutils"
)

func TestFS_Create(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := engineos.New(store)
	require.NoError(t, err)

	filename := "/fs/create/file"
	_, err = fs.Create(filename)
	require.NoError(t, err)

	_, err = store.Get(context.Background(), filename)
	require.NoError(t, err)

	_, err = fs.Create(filename)
	require.NoError(t, err)
}

func TestFS_Open(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := engineos.New(store)
	require.NoError(t, err)

	filename := fmt.Sprintf("/fs/open/file_%d", rand.Int())
	_, err = store.PutString(context.Background(), filename, "")
	require.NoError(t, err)

	_, err = fs.Open(filename)
	require.NoError(t, err)

	_, err = fs.Open("/fs/open/notexists")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_ReadFile(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := engineos.New(store)
	require.NoError(t, err)

	filename := fmt.Sprintf("/fs/read_file/file_%d", rand.Int())
	message := "hello world"
	_, err = store.PutString(context.Background(), filename, message)
	require.NoError(t, err)

	b, err := fs.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, message, string(b))

	_, err = fs.ReadFile("/fs/read_file/notexists")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_Remove(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := engineos.New(store)
	require.NoError(t, err)

	filename := fmt.Sprintf("/fs/remove/file_%d", rand.Int())
	_, err = store.PutString(context.Background(), filename, "")
	require.NoError(t, err)

	err = fs.Remove(filename)
	require.NoError(t, err)

	err = fs.Remove("/fs/remove/notexists")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_Rename(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := engineos.New(store)
	require.NoError(t, err)

	filename := fmt.Sprintf("/fs/rename/file_%d", rand.Int())
	_, err = store.PutString(context.Background(), filename, "")
	require.NoError(t, err)

	newFilename := filename + "_2"
	err = fs.Rename(filename, newFilename)
	require.NoError(t, err)

	_, err = store.Get(context.Background(), newFilename)
	require.NoError(t, err)
}

func TestFS_WriteFile(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := engineos.New(store)
	require.NoError(t, err)

	filename := fmt.Sprintf("/fs/write_file/file_%d", rand.Int())
	message := []byte("hello world")
	err = fs.WriteFile(filename, message, 0)
	require.NoError(t, err)

	s, err := store.GetString(context.Background(), filename)
	require.NoError(t, err)
	require.Equal(t, string(message), s)
}

func TestFS_ReadDir(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := engineos.New(store)
	require.NoError(t, err)

	filename := fmt.Sprintf("/collector_%d.dat", rand.Int())
	_, err = store.PutString(context.Background(), filename, "")
	require.NoError(t, err)

	filename = fmt.Sprintf("/private/data_%d.log", rand.Int())
	_, err = store.PutString(context.Background(), filename, "")
	require.NoError(t, err)

	for range 3 {
		filename := fmt.Sprintf("/documents/file_%d.pdf", rand.Int())
		_, err = store.PutString(context.Background(), filename, "")
		require.NoError(t, err)
	}

	for range 2 {
		filename := fmt.Sprintf("/music/file_%d.mp3", rand.Int())
		_, err = store.PutString(context.Background(), filename, "")
		require.NoError(t, err)
	}

	for range 5 {
		filename := fmt.Sprintf("/private/documents/file_%d.docm", rand.Int())
		_, err = store.PutString(context.Background(), filename, "")
		require.NoError(t, err)
	}

	files, err := fs.ReadDir("/documents")
	require.NoError(t, err)
	require.Len(t, files, 3)
	require.False(t, files[0].IsDir(), "file '%s' should be a regular file", files[0].Name())
	require.False(t, files[1].IsDir(), "file '%s' should be a regular file", files[1].Name())
	require.False(t, files[2].IsDir(), "file '%s' should be a regular file", files[2].Name())

	files, err = fs.ReadDir("/music")
	require.NoError(t, err)
	require.Len(t, files, 2)
	require.False(t, files[0].IsDir(), "file '%s' should be a regular file", files[0].Name())
	require.False(t, files[1].IsDir(), "file '%s' should be a regular file", files[1].Name())

	files, err = fs.ReadDir("/private")
	require.NoError(t, err)
	require.Len(t, files, 2)
	require.False(t, files[0].IsDir(), "file '%s' should be a regular file", files[0].Name())
	require.True(t, files[1].IsDir(), "file '%s' should be a directory", files[1].Name())

	files, err = fs.ReadDir("/")
	require.NoError(t, err)
	require.Len(t, files, 4)
	require.False(t, files[0].IsDir(), "file '%s' should be a regular file", files[0].Name())
	require.True(t, files[1].IsDir(), "file '%s' should be a directory", files[1].Name())
	require.True(t, files[2].IsDir(), "file '%s' should be a directory", files[2].Name())
	require.True(t, files[3].IsDir(), "file '%s' should be a directory", files[3].Name())

	files, err = fs.ReadDir("/documents/")
	require.NoError(t, err)
	require.Len(t, files, 3)
	require.False(t, files[0].IsDir(), "file '%s' should be a regular file", files[0].Name())
	require.False(t, files[1].IsDir(), "file '%s' should be a regular file", files[1].Name())
	require.False(t, files[2].IsDir(), "file '%s' should be a regular file", files[2].Name())

	files, err = fs.ReadDir("documents")
	require.NoError(t, err)
	require.Len(t, files, 0)

	files, err = fs.ReadDir(".")
	require.NoError(t, err)
	require.Len(t, files, 0)

	_, err = fs.ReadDir("")
	require.ErrorIs(t, err, engineos.ErrInvalidFilename)
}
