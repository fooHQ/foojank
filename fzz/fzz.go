package fzz

import (
	"archive/zip"
	"github.com/otiai10/copy"
	"os"
)

const fileExt = "fzz"

func NewFilename(name string) string {
	return name + "." + fileExt
}

func Build(src, dst string) error {
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
