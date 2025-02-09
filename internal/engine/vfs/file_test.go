package vfs_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	engineos "github.com/foohq/foojank/internal/engine/vfs"
	"github.com/foohq/foojank/internal/testutils"
)

func TestFile_Read(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	filename := fmt.Sprintf("/file/read/file_%d", rand.Int())
	message := "hello world"
	_, err := store.PutString(context.Background(), filename, message)
	require.NoError(t, err)

	f := engineos.NewFile(filename, store)
	b := make([]byte, len(message))
	_, err = f.Read(b)
	require.NoError(t, err)
	require.Equal(t, message, string(b))

	f = engineos.NewFile("/file/read/notexists", store)
	b = make([]byte, len(message))
	_, err = f.Read(b)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFile_Write(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	filename := fmt.Sprintf("/file/write/file_%d", rand.Int())
	message := "hello world"
	f := engineos.NewFile(filename, store)
	_, err := f.Write([]byte(message))
	require.NoError(t, err)

	s, err := store.GetString(context.Background(), filename)
	require.NoError(t, err)
	require.Equal(t, message, s)
}
