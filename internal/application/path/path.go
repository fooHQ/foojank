package path

import (
	"errors"
	"strings"
)

type Path struct {
	Repository string
	FilePath   string
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

	if repo == "" || filePath == "" {
		return Path{}, errors.New("repository and path cannot be empty")
	}

	return Path{
		Repository: repo,
		FilePath:   filePath,
	}, nil
}
