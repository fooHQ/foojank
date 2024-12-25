package fzz_test

import (
	"archive/zip"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/foohq/foojank/fzz"
)

func TestBuild(t *testing.T) {
	out := fzz.NewFilename("helo")
	defer os.Remove(out)
	err := fzz.Build("testdata/helo", out)
	assert.NoError(t, err)

	zr, err := zip.OpenReader(out)
	assert.NoError(t, err)
	assert.Equal(t, "main.risor", zr.File[0].Name)
	zr.Close()
}

func TestBuildIsEmpty(t *testing.T) {
	src := "/tmp/aaa"
	err := os.Mkdir(src, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(src)

	out := fzz.NewFilename("empty")
	err = fzz.Build(src, out)
	assert.ErrorIs(t, err, fzz.ErrIsEmpty)
}

func TestBuildMissingMain(t *testing.T) {
	out := fzz.NewFilename("helo")
	err := fzz.Build("testdata/nomain", out)
	assert.ErrorIs(t, err, fzz.ErrMissingMain)
}

func TestBuildInvalidMain(t *testing.T) {
	out := fzz.NewFilename("helo")
	err := fzz.Build("testdata/noregmain", out)
	assert.ErrorIs(t, err, fzz.ErrInvalidMain)
}
