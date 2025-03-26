package nats_test

import (
	"context"
	"fmt"
	"io"
	iofs "io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	natsfs "github.com/foohq/foojank/filesystems/nats"
	"github.com/foohq/foojank/internal/testutils"
)

func TestFS_Create(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	_, err = fs.Create("/test.txt")
	require.NoError(t, err)

	_, err = fs.Stat("/test.txt")
	require.NoError(t, err)
}

func TestFS_Mkdir(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.Mkdir("/dir", 0755)
	require.Equal(t, natsfs.ErrDirectoriesNotSupported, err)
}

func TestFS_MkdirAll(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.MkdirAll("/dir/sub", 0755)
	require.Equal(t, natsfs.ErrDirectoriesNotSupported, err)
}

func TestFS_Open(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test.txt", nil, 0644)
	require.NoError(t, err)

	f, err := fs.Open("/test.txt")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)
}

func TestFS_OpenFile(t *testing.T) {
	t.Parallel()
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	t.Run("CreateWriteOnly", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		require.NoError(t, err)
		defer f.Close()

		n, err := f.Write([]byte("hello world"))
		require.NoError(t, err)
		require.Equal(t, 11, n)

		_, err = fs.Stat(filename)
		require.NoError(t, err)
	})

	t.Run("OpenWriteOnly", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		require.NoError(t, err)
		f.Close()

		f, err = fs.OpenFile(filename, os.O_WRONLY, 0)
		require.NoError(t, err)
		defer f.Close()

		n, err := f.Write([]byte("hello world"))
		require.NoError(t, err)
		require.Equal(t, 11, n)
	})

	t.Run("CreateReadOnly", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		f, err := fs.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
		require.NoError(t, err)
		defer f.Close()

		n, err := f.Read([]byte("hello world"))
		require.ErrorIs(t, err, io.EOF)
		require.Equal(t, 0, n)

		_, err = fs.Stat(filename)
		require.NoError(t, err)
	})

	t.Run("OpenReadOnly", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		err := fs.WriteFile(filename, []byte("hello world"), 0644)
		require.NoError(t, err)

		f, err := fs.OpenFile(filename, os.O_RDONLY, 0)
		require.NoError(t, err)
		defer f.Close()

		b := make([]byte, 11)
		n, err := f.Read(b)
		require.NoError(t, err)
		require.Equal(t, 11, n)
		require.Equal(t, []byte("hello world"), b)
	})

	t.Run("OpenNonExistent", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		_, err := fs.OpenFile(filename, os.O_WRONLY, 0)
		require.ErrorIs(t, err, os.ErrNotExist)

		_, err = fs.OpenFile(filename, os.O_RDONLY, 0)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("CreateReadWrite", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		f, err := fs.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
		require.NoError(t, err)
		defer f.Close()

		n, err := f.Write([]byte("hello world"))
		require.NoError(t, err)
		require.Equal(t, 11, n)

		b := make([]byte, 11)
		_, err = f.Read(b)
		// risoros.File does not use fs.Seeker interface, so there is no way to reset internal offset.
		require.ErrorIs(t, err, io.EOF)

		_, err = fs.Stat(filename)
		require.NoError(t, err)
	})

	t.Run("WriteBadDescriptor", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		f, err := fs.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
		require.NoError(t, err)
		defer f.Close()

		_, err = f.Write([]byte("hello world"))
		require.ErrorIs(t, err, natsfs.ErrBadDescriptor)
	})

	t.Run("ReadBadDescriptor", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		require.NoError(t, err)
		defer f.Close()

		b := make([]byte, 11)
		_, err = f.Read(b)
		require.ErrorIs(t, err, natsfs.ErrBadDescriptor)
	})

	t.Run("OpenWriteRootDirectory", func(t *testing.T) {
		filename := "/"
		_, err := fs.OpenFile(filename, os.O_RDWR, 0)
		require.ErrorIs(t, err, natsfs.ErrIsDirectory)

		_, err = fs.OpenFile(filename, os.O_WRONLY, 0)
		require.ErrorIs(t, err, natsfs.ErrIsDirectory)
	})

	t.Run("OpenWriteDirectory", func(t *testing.T) {
		filename := fmt.Sprintf("dir_%d/file", rand.Int())
		err := fs.WriteFile(filename, nil, 0644)
		require.NoError(t, err)

		dirname := filepath.Dir(filename)
		_, err = fs.OpenFile(dirname, os.O_RDWR, 0)
		require.ErrorIs(t, err, natsfs.ErrIsDirectory)
	})

	t.Run("OpenReadRootDirectory", func(t *testing.T) {
		filename := "/"
		f, err := fs.OpenFile(filename, os.O_RDONLY, 0)
		require.NoError(t, err)
		defer f.Close()

		info, err := f.Stat()
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("OpenReadDirectory", func(t *testing.T) {
		filename := fmt.Sprintf("dir_%d", rand.Int())
		err := fs.WriteFile(filename, nil, 0644)
		require.NoError(t, err)

		dirname := filepath.Dir(filename)
		f, err := fs.OpenFile(dirname, os.O_RDONLY, 0)
		require.NoError(t, err)
		defer f.Close()

		info, err := f.Stat()
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})
}

func TestFS_ReadFile(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test.txt", []byte("content"), 0644)
	require.NoError(t, err)

	data, err := fs.ReadFile("/test.txt")
	require.NoError(t, err)
	require.Equal(t, "content", string(data))

	_, err = fs.ReadFile("/missing.txt")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_Remove(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test.txt", nil, 0644)
	require.NoError(t, err)

	_, err = fs.Stat("/test.txt")
	require.NoError(t, err)

	err = fs.Remove("/test.txt")
	require.NoError(t, err)

	_, err = fs.Stat("/test.txt")
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = store.GetBytes(context.Background(), "/test.txt")
	require.ErrorIs(t, err, jetstream.ErrObjectNotFound)
}

func TestFS_RemoveAll(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test/a.txt", []byte("a"), 0644)
	require.NoError(t, err)
	err = fs.WriteFile("/test/b.txt", []byte("b"), 0644)
	require.NoError(t, err)

	err = fs.RemoveAll("/test")
	require.NoError(t, err)

	_, err = fs.Stat("/test/a.txt")
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = fs.Stat("/test/b.txt")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_Rename(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/old.txt", []byte("content"), 0644)
	require.NoError(t, err)

	err = fs.Rename("/old.txt", "/new.txt")
	require.NoError(t, err)

	_, err = fs.Stat("/old.txt")
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = fs.Stat("/new.txt")
	require.NoError(t, err)
	data, err := fs.ReadFile("/new.txt")
	require.NoError(t, err)
	require.Equal(t, "content", string(data))
}

func TestFS_Stat(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test.txt", []byte("content"), 0644)
	require.NoError(t, err)

	info, err := fs.Stat("/test.txt")
	require.NoError(t, err)
	require.Equal(t, "test.txt", info.Name())

	_, err = fs.Stat("/missing.txt")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_Symlink(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.Symlink("/target.txt", "/link.txt")
	require.Equal(t, natsfs.ErrSymlinksNotSupported, err)
}

func TestFS_WriteFile(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test.txt", []byte("content"), 0644)
	require.NoError(t, err)

	_, err = fs.Stat("/test.txt")
	require.NoError(t, err)
	data, err := fs.ReadFile("/test.txt")
	require.NoError(t, err)
	require.Equal(t, "content", string(data))
}

func TestFS_ReadDir(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test/a.txt", []byte("a"), 0644)
	require.NoError(t, err)
	err = fs.WriteFile("/test/b.txt", []byte("b"), 0644)
	require.NoError(t, err)

	entries, err := fs.ReadDir("/test")
	require.NoError(t, err)
	require.Len(t, entries, 2)

	names := map[string]bool{"a.txt": false, "b.txt": false}
	for _, e := range entries {
		names[e.Name()] = true
	}
	for name, found := range names {
		require.True(t, found, name)
	}
}

func TestFS_WalkDir(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.WriteFile("/test/a.txt", []byte("a"), 0644)
	require.NoError(t, err)
	err = fs.WriteFile("/test/sub/b.txt", []byte("b"), 0644)
	require.NoError(t, err)

	paths := make(map[string]bool)
	err = fs.WalkDir("/test", func(path string, d iofs.DirEntry, err error) error {
		require.NoError(t, err)
		paths[path] = true
		return nil
	})
	require.NoError(t, err)

	expected := []string{"/test", "/test/a.txt", "/test/sub", "/test/sub/b.txt"}
	for _, p := range expected {
		require.True(t, paths[p])
	}
}

func TestFS_Close(t *testing.T) {
	_, nc := testutils.NewNatsServerAndConnection(t)
	store := testutils.NewNatsObjectStore(t, nc)

	fs, err := natsfs.NewFS(store)
	require.NoError(t, err)
	defer fs.Close()

	err = fs.Close()
	require.NoError(t, err)

	_, err = fs.Open("/test.txt")
	require.Error(t, err) // Should fail due to canceled context
}
