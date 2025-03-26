package mem_test

import (
	"fmt"
	"io"
	iofs "io/fs"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	memfs "github.com/foohq/foojank/filesystems/mem"
)

func TestFS_Create(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	f, err := fs.Create("test.txt")
	require.NoError(t, err)
	defer f.Close()

	info, err := fs.Stat("test.txt")
	require.NoError(t, err)
	require.False(t, info.IsDir())
}

func TestFS_Mkdir(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	err = fs.Mkdir("dir", 0755)
	require.NoError(t, err)

	info, err := fs.Stat("dir")
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestFS_MkdirAll(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	err = fs.MkdirAll("a/b/c", 0755)
	require.NoError(t, err)

	info, err := fs.Stat("a/b/c")
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestFS_Open(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	_, err = fs.Create("test.txt")
	require.NoError(t, err)

	f, err := fs.Open("test.txt")
	require.NoError(t, err)
	defer f.Close()
}

func TestFS_OpenFile(t *testing.T) {
	t.Parallel()
	fs, err := memfs.NewFS()
	require.NoError(t, err)

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

		b := make([]byte, 11)
		n, err := f.Read(b)
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
		require.ErrorIs(t, err, memfs.ErrBadDescriptor)
	})

	t.Run("ReadBadDescriptor", func(t *testing.T) {
		filename := fmt.Sprintf("file_%d", rand.Int())
		f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		require.NoError(t, err)
		defer f.Close()

		b := make([]byte, 11)
		_, err = f.Read(b)
		require.ErrorIs(t, err, memfs.ErrBadDescriptor)
	})

	t.Run("OpenWriteRootDirectory", func(t *testing.T) {
		filename := "/"
		_, err := fs.OpenFile(filename, os.O_RDWR, 0)
		require.ErrorIs(t, err, memfs.ErrIsDirectory)

		_, err = fs.OpenFile(filename, os.O_WRONLY, 0)
		require.ErrorIs(t, err, memfs.ErrIsDirectory)
	})

	t.Run("OpenWriteDirectory", func(t *testing.T) {
		filename := fmt.Sprintf("dir_%d", rand.Int())
		err := fs.MkdirAll(filename, 0755)
		require.NoError(t, err)

		_, err = fs.OpenFile(filename, os.O_RDWR, 0)
		require.ErrorIs(t, err, memfs.ErrIsDirectory)
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
		err := fs.MkdirAll(filename, 0755)
		require.NoError(t, err)

		f, err := fs.OpenFile(filename, os.O_RDONLY, 0)
		require.NoError(t, err)
		defer f.Close()

		info, err := f.Stat()
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})
}

func TestFS_ReadFile(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	err = fs.WriteFile("test.txt", []byte("hello"), 0644)
	require.NoError(t, err)

	data, err := fs.ReadFile("test.txt")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)
}

func TestFS_Remove(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	_, err = fs.Create("test.txt")
	require.NoError(t, err)

	err = fs.Remove("test.txt")
	require.NoError(t, err)

	_, err = fs.Stat("test.txt")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_RemoveAll(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	err = fs.MkdirAll("a/b/c", 0755)
	require.NoError(t, err)

	err = fs.RemoveAll("/")
	require.NoError(t, err)

	_, err = fs.Stat("/a")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFS_Rename(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	_, err = fs.Create("old.txt")
	require.NoError(t, err)

	err = fs.Rename("old.txt", "new.txt")
	require.NoError(t, err)

	_, err = fs.Stat("old.txt")
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = fs.Stat("new.txt")
	require.NoError(t, err)
}

func TestFS_Stat(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	_, err = fs.Create("test.txt")
	require.NoError(t, err)

	info, err := fs.Stat("test.txt")
	require.NoError(t, err)
	require.Equal(t, "test.txt", info.Name())
}

func TestFS_Symlink(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	_, err = fs.Create("target.txt")
	require.NoError(t, err)

	err = fs.Symlink("target.txt", "link.txt")
	require.NoError(t, err)

	info, err := fs.Stat("link.txt")
	require.NoError(t, err)
	require.Equal(t, "link.txt", info.Name())
}

func TestFS_WriteFile(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	err = fs.WriteFile("test.txt", []byte("hello"), 0644)
	require.NoError(t, err)

	data, err := fs.ReadFile("test.txt")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)
}

func TestFS_ReadDir(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	err = fs.Mkdir("dir", 0755)
	require.NoError(t, err)

	_, err = fs.Create("dir/file.txt")
	require.NoError(t, err)

	entries, err := fs.ReadDir("dir")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "file.txt", entries[0].Name())
}

func TestFS_WalkDir(t *testing.T) {
	fs, err := memfs.NewFS()
	require.NoError(t, err)

	err = fs.MkdirAll("a/b", 0755)
	require.NoError(t, err)

	_, err = fs.Create("a/file.txt")
	require.NoError(t, err)

	var paths []string
	err = fs.WalkDir("a", func(path string, d iofs.DirEntry, err error) error {
		require.NoError(t, err)
		paths = append(paths, path)
		return nil
	})
	require.NoError(t, err)

	expected := []string{"a", "a/b", "a/file.txt"}
	require.ElementsMatch(t, expected, paths)
}
