package fzz

import (
	"archive/zip"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBuild(t *testing.T) {
	out := NewFilename("helo")
	defer os.Remove(out)
	err := Build("testdata/helo", out)
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

	out := NewFilename("empty")
	err = Build(src, out)
	assert.ErrorIs(t, err, ErrIsEmpty)
}

func TestBuildMissingMain(t *testing.T) {
	out := NewFilename("helo")
	err := Build("testdata/nomain", out)
	assert.ErrorIs(t, err, ErrMissingMain)
}

func TestBuildInvalidMain(t *testing.T) {
	out := NewFilename("helo")
	err := Build("testdata/noregmain", out)
	assert.ErrorIs(t, err, ErrInvalidMain)
}
