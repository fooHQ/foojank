package fzz

import (
	"archive/zip"
	"errors"
	"github.com/otiai10/copy"
	"io"
	"os"
)

var (
	ErrIsEmpty = errors.New("directory is empty")
)

const fileExt = "fzz"

func NewFilename(name string) string {
	return name + "." + fileExt
}

func Build(src, dst string) error {
	empty, err := isEmpty(src)
	if err != nil {
		return err
	}

	if empty {
		return ErrIsEmpty
	}

	tmpDir, err := os.MkdirTemp(".", "fzz*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	err = copy.Copy(src, tmpDir)
	if err != nil {
		return err
	}

	f, err := os.CreateTemp(".", "fzz*."+fileExt)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	zw := zip.NewWriter(f)
	defer zw.Close()

	err = zw.AddFS(os.DirFS(tmpDir))
	if err != nil {
		return err
	}

	err = zw.Close()
	if err != nil {
		return err
	}

	err = os.Rename(f.Name(), dst)
	if err != nil {
		return err
	}

	return nil
}

func isEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}
