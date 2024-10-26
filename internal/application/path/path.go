package path

import (
	"errors"
	"path/filepath"
	"strings"
)

type Path struct {
	Repository string
	FilePath   string
}

func (p Path) IsLocal() bool {
	return p.Repository == ""
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
	return strings.Join([]string{p.Repository, p.FilePath}, ":")
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
	repo := strings.TrimSpace(parts[0])
	filePath := strings.TrimSpace(parts[1])
	// TODO: cleanup filePath (it must always be an absolute path!)

	if repo == "" || filePath == "" {
		return Path{}, errors.New("repository and path cannot be empty")
	}

	return Path{
		Repository: repo,
		FilePath:   filePath,
	}, nil
}
