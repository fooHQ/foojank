package path

import (
	"errors"
	"path/filepath"
	"strings"
)

type Path struct {
	Storage  string
	FilePath string
}

func (p Path) IsLocal() bool {
	return p.Storage == ""
}

func (p Path) IsDir() bool {
	return strings.HasSuffix(p.FilePath, "/")
}

func (p Path) Base() string {
	return filepath.Base(p.FilePath)
}

func (p Path) String() string {
	if p.IsLocal() {
		return p.FilePath
	}
	return strings.Join([]string{p.Storage, p.FilePath}, ":")
}

func Parse(input string) (Path, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Path{}, errors.New("path cannot be empty")
	}

	if !strings.Contains(input, ":") {
		return Path{
			FilePath: input,
		}, nil
	}

	parts := strings.SplitN(input, ":", 2)
	storage := strings.TrimSpace(parts[0])
	filePath := strings.TrimSpace(parts[1])

	if storage == "" || filePath == "" {
		return Path{}, errors.New("storage name and path cannot be empty")
	}

	return Path{
		Storage:  storage,
		FilePath: filePath,
	}, nil
}
