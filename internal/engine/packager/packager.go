package packager

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

var (
	ErrIsEmpty     = errors.New("directory is empty")
	ErrInvalidMain = errors.New("main file is not a regular file")
	ErrMissingMain = errors.New("main file is missing")
)

const fileExt = "fzz"

func NewFilename(name string) string {
	if strings.HasSuffix(name, fileExt) {
		return name
	}
	return name + "." + fileExt
}

func Build(src, dst string) error {
	err := isEmpty(src)
	if err != nil {
		return err
	}

	err = isMain(src)
	if err != nil {
		return err
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

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(f.Name(), dst)
	if err != nil {
		return err
	}

	return nil
}

func isEmpty(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return ErrIsEmpty
	}

	return err
}

func isMain(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	files, err := f.Readdirnames(-1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return err
		}
		return ErrMissingMain
	}

	for _, name := range files {
		if name != "main.risor" && name != "main.rsr" {
			continue
		}

		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return ErrInvalidMain
		}

		return nil
	}

	return ErrMissingMain
}
