package packager_test

import (
	"archive/zip"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/engine/packager"
)

func TestBuild(t *testing.T) {
	out := packager.NewFilename("helo")
	defer os.Remove(out)
	err := packager.Build("testdata/helo", out)
	require.NoError(t, err)

	zr, err := zip.OpenReader(out)
	require.NoError(t, err)
	require.Equal(t, "main.risor", zr.File[0].Name)
	zr.Close()
}

func TestBuildIsEmpty(t *testing.T) {
	src, err := os.MkdirTemp(os.TempDir(), "fzz")
	require.NoError(t, err)
	defer os.RemoveAll(src)

	out := packager.NewFilename("empty")
	err = packager.Build(src, out)
	require.ErrorIs(t, err, packager.ErrIsEmpty)
}

func TestBuildMissingMain(t *testing.T) {
	out := packager.NewFilename("helo")
	err := packager.Build("testdata/nomain", out)
	require.ErrorIs(t, err, packager.ErrMissingMain)
}

func TestBuildInvalidMain(t *testing.T) {
	out := packager.NewFilename("helo")
	err := packager.Build("testdata/noregmain", out)
	require.ErrorIs(t, err, packager.ErrInvalidMain)
}
