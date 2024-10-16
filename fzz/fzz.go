package fzz

import (
	"archive/zip"
	"github.com/otiai10/copy"
	"os"
)

const fileExt = "fzz"

func Build(src, name string) error {
	tmpDir, err := os.MkdirTemp(".", "fzz*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	err = copy.Copy(src, tmpDir)
	if err != nil {
		return err
	}

	f, err := os.CreateTemp(".", "fzz*.fzz")
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

	name = name + "." + fileExt
	err = os.Rename(f.Name(), name)
	if err != nil {
		return err
	}

	return nil
}
