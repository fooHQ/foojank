package fzz_test

import (
	"archive/zip"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/fzz"
)

func TestBuild(t *testing.T) {
	out := fzz.NewFilename("helo")
	defer os.Remove(out)
	err := fzz.Build("testdata/helo", out)
	require.NoError(t, err)

	zr, err := zip.OpenReader(out)
	require.NoError(t, err)
	require.Equal(t, "main.risor", zr.File[0].Name)
	zr.Close()
}

func TestBuildIsEmpty(t *testing.T) {
	src := "/tmp/aaa"
	err := os.Mkdir(src, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(src)

	out := fzz.NewFilename("empty")
	err = fzz.Build(src, out)
	require.ErrorIs(t, err, fzz.ErrIsEmpty)
}

func TestBuildMissingMain(t *testing.T) {
	out := fzz.NewFilename("helo")
	err := fzz.Build("testdata/nomain", out)
	require.ErrorIs(t, err, fzz.ErrMissingMain)
}

func TestBuildInvalidMain(t *testing.T) {
	out := fzz.NewFilename("helo")
	err := fzz.Build("testdata/noregmain", out)
	require.ErrorIs(t, err, fzz.ErrInvalidMain)
}
