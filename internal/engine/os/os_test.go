package os_test

import (
	"context"
	"testing"

	risoros "github.com/risor-io/risor/os"
	"github.com/stretchr/testify/require"

	engineos "github.com/foohq/foojank/internal/engine/os"
	"github.com/foohq/foojank/internal/testutils"
)

// TODO: add the rest of the tests!

func TestOS_Args(t *testing.T) {
	args := []string{
		"first",
		"second",
		"third",
	}
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithArgs(args),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	actualArgs := o.Args()
	require.Equal(t, args, actualArgs)
}

func TestOS_Create(t *testing.T) {
	testCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(testCh)
	fsFile := testutils.NewURIHandler(testCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.CreateResult
	}{
		{
			input: "test://private/",
			result: testutils.CreateResult{
				Name: "/",
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.CreateResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.CreateResult{
				Name: "/form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.CreateResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/",
			result: testutils.CreateResult{
				Name: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.CreateResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.CreateResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.CreateResult{
				Name: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.CreateResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.CreateResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.CreateResult{
				Name: "/foojank/data/form.txt",
			},
		},
	}

	for i, test := range tests {
		_, err := o.Create(test.input)
		require.NoError(t, err, "test %d", i)

		result := <-testCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Mkdir(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.MkdirResult
	}{
		{
			input: "test://private/",
			result: testutils.MkdirResult{
				Name: "/",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/form",
			result: testutils.MkdirResult{
				Name: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../form",
			result: testutils.MkdirResult{
				Name: "/form",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.MkdirResult{
				Name: "/form",
				Perm: 0777,
			},
		},
		{
			input: "/",
			result: testutils.MkdirResult{
				Name: "/",
				Perm: 0777,
			},
		},
		{
			input: "/data/form",
			result: testutils.MkdirResult{
				Name: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "/data/../form",
			result: testutils.MkdirResult{
				Name: "/form",
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form",
			result: testutils.MkdirResult{
				Name: "/form",
				Perm: 0777,
			},
		},
		{
			input: "file:///data/form",
			result: testutils.MkdirResult{
				Name: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "../data/form",
			result: testutils.MkdirResult{
				Name: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.MkdirResult{
				Name: "/foojank/data/form.txt",
				Perm: 0777,
			},
		},
	}

	for _, test := range tests {
		err := o.Mkdir(test.input, 0777)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_MkdirAll(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.MkdirAllResult
	}{
		{
			input: "test://private/",
			result: testutils.MkdirAllResult{
				Path: "/",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/form",
			result: testutils.MkdirAllResult{
				Path: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../form",
			result: testutils.MkdirAllResult{
				Path: "/form",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.MkdirAllResult{
				Path: "/form",
				Perm: 0777,
			},
		},
		{
			input: "/",
			result: testutils.MkdirAllResult{
				Path: "/",
				Perm: 0777,
			},
		},
		{
			input: "/data/form",
			result: testutils.MkdirAllResult{
				Path: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "/data/../form",
			result: testutils.MkdirAllResult{
				Path: "/form",
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form",
			result: testutils.MkdirAllResult{
				Path: "/form",
				Perm: 0777,
			},
		},
		{
			input: "file:///data/form",
			result: testutils.MkdirAllResult{
				Path: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "../data/form",
			result: testutils.MkdirAllResult{
				Path: "/data/form",
				Perm: 0777,
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.MkdirAllResult{
				Path: "/foojank/data/form.txt",
				Perm: 0777,
			},
		},
	}

	for _, test := range tests {
		err := o.MkdirAll(test.input, 0777)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Open(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.OpenResult
	}{
		{
			input: "test://private/",
			result: testutils.OpenResult{
				Name: "/",
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.OpenResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.OpenResult{
				Name: "/form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.OpenResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/",
			result: testutils.OpenResult{
				Name: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.OpenResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.OpenResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.OpenResult{
				Name: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.OpenResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.OpenResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.OpenResult{
				Name: "/foojank/data/form.txt",
			},
		},
	}

	for _, test := range tests {
		_, err := o.Open(test.input)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_OpenFile(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.OpenFileResult
	}{
		{
			input: "test://private/",
			result: testutils.OpenFileResult{
				Name: "/",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.OpenFileResult{
				Name: "/data/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.OpenFileResult{
				Name: "/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.OpenFileResult{
				Name: "/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "/",
			result: testutils.OpenFileResult{
				Name: "/",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.OpenFileResult{
				Name: "/data/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.OpenFileResult{
				Name: "/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.OpenFileResult{
				Name: "/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.OpenFileResult{
				Name: "/data/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.OpenFileResult{
				Name: "/data/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.OpenFileResult{
				Name: "/foojank/data/form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
	}

	for _, test := range tests {
		_, err := o.OpenFile(test.input, 1313, 0777)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_ReadFile(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.ReadFileResult
	}{
		{
			input: "test://private/",
			result: testutils.ReadFileResult{
				Name: "/",
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.ReadFileResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.ReadFileResult{
				Name: "/form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.ReadFileResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/",
			result: testutils.ReadFileResult{
				Name: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.ReadFileResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.ReadFileResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.ReadFileResult{
				Name: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.ReadFileResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.ReadFileResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.ReadFileResult{
				Name: "/foojank/data/form.txt",
			},
		},
	}

	for _, test := range tests {
		_, err := o.ReadFile(test.input)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Remove(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.RemoveResult
	}{
		{
			input: "test://private/",
			result: testutils.RemoveResult{
				Name: "/",
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.RemoveResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.RemoveResult{
				Name: "/form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.RemoveResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/",
			result: testutils.RemoveResult{
				Name: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.RemoveResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.RemoveResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.RemoveResult{
				Name: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.RemoveResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.RemoveResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.RemoveResult{
				Name: "/foojank/data/form.txt",
			},
		},
	}

	for _, test := range tests {
		err := o.Remove(test.input)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_RemoveAll(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.RemoveAllResult
	}{
		{
			input: "test://private/",
			result: testutils.RemoveAllResult{
				Path: "/",
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.RemoveAllResult{
				Path: "/data/form.txt",
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.RemoveAllResult{
				Path: "/form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.RemoveAllResult{
				Path: "/form.txt",
			},
		},
		{
			input: "/",
			result: testutils.RemoveAllResult{
				Path: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.RemoveAllResult{
				Path: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.RemoveAllResult{
				Path: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.RemoveAllResult{
				Path: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.RemoveAllResult{
				Path: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.RemoveAllResult{
				Path: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.RemoveAllResult{
				Path: "/foojank/data/form.txt",
			},
		},
	}

	for _, test := range tests {
		err := o.RemoveAll(test.input)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Rename(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		src    string
		dst    string
		result testutils.RenameResult
	}{
		{
			src: "test://private/foo.txt",
			dst: "test://private/bar.txt",
			result: testutils.RenameResult{
				OldPath: "/foo.txt",
				NewPath: "/bar.txt",
			},
		},
		{
			src: "/private/foo.txt",
			dst: "/private/bar.txt",
			result: testutils.RenameResult{
				OldPath: "/private/foo.txt",
				NewPath: "/private/bar.txt",
			},
		},
		{
			src: "/private/foo.txt",
			dst: "../bar.txt",
			result: testutils.RenameResult{
				OldPath: "/private/foo.txt",
				NewPath: "/bar.txt",
			},
		},
		{
			src: "./private/foo.txt",
			dst: "bar.txt",
			result: testutils.RenameResult{
				OldPath: "/foojank/private/foo.txt",
				NewPath: "/foojank/bar.txt",
			},
		},
		{
			src: "../foo.txt",
			dst: "./bar.txt",
			result: testutils.RenameResult{
				OldPath: "/foo.txt",
				NewPath: "/foojank/bar.txt",
			},
		},
	}

	for _, test := range tests {
		err := o.Rename(test.src, test.dst)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Rename_ErrCrossingFSBoundaries(t *testing.T) {
	fsPrivate := testutils.NewURIHandler(nil)
	fsFile := testutils.NewURIHandler(nil)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		src string
		dst string
	}{
		{
			src: "test://private/foo.txt",
			dst: "./private/bar.txt",
		},
		{
			src: "./private/bar.txt",
			dst: "test://private/foo.txt",
		},
	}

	for _, test := range tests {
		err := o.Rename(test.src, test.dst)
		require.ErrorIs(t, err, engineos.ErrCrossingFSBoundaries)
	}
}

func TestOS_Stat(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.StatResult
	}{
		{
			input: "test://private/",
			result: testutils.StatResult{
				Name: "/",
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.StatResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.StatResult{
				Name: "/form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.StatResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/",
			result: testutils.StatResult{
				Name: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.StatResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.StatResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.StatResult{
				Name: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.StatResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.StatResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.StatResult{
				Name: "/foojank/data/form.txt",
			},
		},
	}

	for _, test := range tests {
		_, err := o.Stat(test.input)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Symlink(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		src    string
		dst    string
		result testutils.SymlinkResult
	}{
		{
			src: "test://private/foo.txt",
			dst: "test://private/bar.txt",
			result: testutils.SymlinkResult{
				OldName: "/foo.txt",
				NewName: "/bar.txt",
			},
		},
		{
			src: "/private/foo.txt",
			dst: "/private/bar.txt",
			result: testutils.SymlinkResult{
				OldName: "/private/foo.txt",
				NewName: "/private/bar.txt",
			},
		},
		{
			src: "/private/foo.txt",
			dst: "../bar.txt",
			result: testutils.SymlinkResult{
				OldName: "/private/foo.txt",
				NewName: "/bar.txt",
			},
		},
		{
			src: "./private/foo.txt",
			dst: "bar.txt",
			result: testutils.SymlinkResult{
				OldName: "/foojank/private/foo.txt",
				NewName: "/foojank/bar.txt",
			},
		},
		{
			src: "../foo.txt",
			dst: "./bar.txt",
			result: testutils.SymlinkResult{
				OldName: "/foo.txt",
				NewName: "/foojank/bar.txt",
			},
		},
	}

	for _, test := range tests {
		err := o.Symlink(test.src, test.dst)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Symlink_ErrCrossingFSBoundaries(t *testing.T) {
	fsPrivate := testutils.NewURIHandler(nil)
	fsFile := testutils.NewURIHandler(nil)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		src string
		dst string
	}{
		{
			src: "test://private/foo.txt",
			dst: "./private/bar.txt",
		},
		{
			src: "./private/bar.txt",
			dst: "test://private/foo.txt",
		},
	}

	for _, test := range tests {
		err := o.Symlink(test.src, test.dst)
		require.ErrorIs(t, err, engineos.ErrCrossingFSBoundaries)
	}
}

func TestOS_WriteFile(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.WriteFileResult
	}{
		{
			input: "test://private/",
			result: testutils.WriteFileResult{
				Name: "/",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/form.txt",
			result: testutils.WriteFileResult{
				Name: "/data/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../form.txt",
			result: testutils.WriteFileResult{
				Name: "/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.WriteFileResult{
				Name: "/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "/",
			result: testutils.WriteFileResult{
				Name: "/",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.WriteFileResult{
				Name: "/data/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.WriteFileResult{
				Name: "/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.WriteFileResult{
				Name: "/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.WriteFileResult{
				Name: "/data/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.WriteFileResult{
				Name: "/data/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.WriteFileResult{
				Name: "/foojank/data/form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
	}

	for _, test := range tests {
		err := o.WriteFile(test.input, []byte("test"), 0777)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_ReadDir(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.ReadDirResult
	}{
		{
			input: "test://private/",
			result: testutils.ReadDirResult{
				Name: "/",
			},
		},
		{
			input: "test://private/data/form",
			result: testutils.ReadDirResult{
				Name: "/data/form",
			},
		},
		{
			input: "test://private/data/../form",
			result: testutils.ReadDirResult{
				Name: "/form",
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.ReadDirResult{
				Name: "/form",
			},
		},
		{
			input: "/",
			result: testutils.ReadDirResult{
				Name: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.ReadDirResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.ReadDirResult{
				Name: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.ReadDirResult{
				Name: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.ReadDirResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.ReadDirResult{
				Name: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.ReadDirResult{
				Name: "/foojank/data/form.txt",
			},
		},
	}

	for _, test := range tests {
		_, err := o.ReadDir(test.input)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_WalkDir(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewURIHandler(resultCh)
	fsFile := testutils.NewURIHandler(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURIHandler("file", fsFile),
		engineos.WithURIHandler("test", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.WalkDirResult
	}{
		{
			input: "test://private/",
			result: testutils.WalkDirResult{
				Root: "/",
				Fn:   nil,
			},
		},
		{
			input: "test://private/data/form",
			result: testutils.WalkDirResult{
				Root: "/data/form",
				Fn:   nil,
			},
		},
		{
			input: "test://private/data/../form",
			result: testutils.WalkDirResult{
				Root: "/form",
				Fn:   nil,
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.WalkDirResult{
				Root: "/form",
				Fn:   nil,
			},
		},
		{
			input: "/",
			result: testutils.WalkDirResult{
				Root: "/",
			},
		},
		{
			input: "/data/form.txt",
			result: testutils.WalkDirResult{
				Root: "/data/form.txt",
			},
		},
		{
			input: "/data/../form.txt",
			result: testutils.WalkDirResult{
				Root: "/form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.WalkDirResult{
				Root: "/form.txt",
			},
		},
		{
			input: "file:///data/form.txt",
			result: testutils.WalkDirResult{
				Root: "/data/form.txt",
			},
		},
		{
			input: "../data/form.txt",
			result: testutils.WalkDirResult{
				Root: "/data/form.txt",
			},
		},
		{
			input: "./data/form.txt",
			result: testutils.WalkDirResult{
				Root: "/foojank/data/form.txt",
			},
		},
	}

	for _, test := range tests {
		err := o.WalkDir(test.input, nil)
		require.NoError(t, err)

		result := <-resultCh
		require.Equal(t, test.result, result)
	}
}
