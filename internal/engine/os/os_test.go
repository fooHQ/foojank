package os_test

import (
	"context"
	"testing"

	risoros "github.com/risor-io/risor/os"
	"github.com/stretchr/testify/require"

	engineos "github.com/foohq/foojank/internal/engine/os"
	"github.com/foohq/foojank/internal/testutils"
)

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
	fsPrivate := testutils.NewFS(testCh)
	fsFile := testutils.NewFS(testCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.CreateResult{
				Name: "/data/../../form.txt",
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.CreateResult{
				Name: "/data/../../form.txt",
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

	for _, test := range tests {
		_, err := o.Create(test.input)
		require.NoError(t, err)

		result := <-testCh
		require.Equal(t, test.result, result)
	}
}

func TestOS_Mkdir(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.MkdirResult{
				Name: "/data/../../form",
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
				Name: "/data/../form",
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form",
			result: testutils.MkdirResult{
				Name: "/data/../../form",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Path: "/data/../form",
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.MkdirAllResult{
				Path: "/data/../../form",
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
				Path: "/data/../form",
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form",
			result: testutils.MkdirAllResult{
				Path: "/data/../../form",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.OpenResult{
				Name: "/data/../../form.txt",
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.OpenResult{
				Name: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.OpenFileResult{
				Name: "/data/../../form.txt",
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
				Name: "/data/../form.txt",
				Flag: 1313,
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.OpenFileResult{
				Name: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.ReadFileResult{
				Name: "/data/../../form.txt",
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.ReadFileResult{
				Name: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.RemoveResult{
				Name: "/data/../../form.txt",
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.RemoveResult{
				Name: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Path: "/data/../form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.RemoveAllResult{
				Path: "/data/../../form.txt",
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
				Path: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.RemoveAllResult{
				Path: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.RenameResult
	}{
		{
			input: "test://private/",
			result: testutils.RenameResult{
				OldPath: "/old.txt",
				NewPath: "/new.txt",
			},
		},
	}

	for _, test := range tests {
		err := o.Rename(test.input, "/new.txt")
		require.Error(t, err)
		/*
			// Check result once the Rename is implemented
			result := <-resultCh
			require.Equal(t, test.result, result)
		*/
	}
}

func TestOS_Stat(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.StatResult{
				Name: "/data/../../form.txt",
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.StatResult{
				Name: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
	)
	o, ok := risoros.GetOS(osCtx)
	require.True(t, ok)

	tests := []struct {
		input  string
		result testutils.SymlinkResult
	}{
		{
			input: "test://private/",
			result: testutils.SymlinkResult{
				OldName: "/old.txt",
				NewName: "/new.txt",
			},
		},
	}

	for _, test := range tests {
		err := o.Symlink(test.input, "/new.txt")
		require.Error(t, err)
		/*
			// Check result once the Rename is implemented
			result := <-resultCh
			require.Equal(t, test.result, result)
		*/
	}
}

func TestOS_WriteFile(t *testing.T) {
	resultCh := make(chan any, 1)
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "test://private/data/../../form.txt",
			result: testutils.WriteFileResult{
				Name: "/data/../../form.txt",
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
				Name: "/data/../form.txt",
				Data: []byte("test"),
				Perm: 0777,
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.WriteFileResult{
				Name: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Name: "/data/../form",
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.ReadDirResult{
				Name: "/data/../../form",
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
				Name: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.ReadDirResult{
				Name: "/data/../../form.txt",
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
	fsPrivate := testutils.NewFS(resultCh)
	fsFile := testutils.NewFS(resultCh)
	osCtx := engineos.NewContext(context.Background(),
		engineos.WithWorkDir("/foojank"),
		engineos.WithURLHandler("file", "", fsFile),
		engineos.WithURLHandler("test", "private", fsPrivate),
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
				Root: "/data/../form",
				Fn:   nil,
			},
		},
		{
			input: "test://private/data/../../form",
			result: testutils.WalkDirResult{
				Root: "/data/../../form",
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
				Root: "/data/../form.txt",
			},
		},
		{
			input: "/data/../../form.txt",
			result: testutils.WalkDirResult{
				Root: "/data/../../form.txt",
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
