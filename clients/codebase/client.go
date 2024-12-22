package codebase

import (
	"os"
	"path/filepath"
)

type Client struct {
	baseDir string
}

func New(path string) *Client {
	return &Client{
		baseDir: path,
	}
}

func (c *Client) GetScript(name string) (string, error) {
	scriptsDir := filepath.Join(c.baseDir, "scripts", name)
	_, err := os.ReadDir(scriptsDir)
	if err != nil {
		return "", err
	}

	return scriptsDir, nil
}

func (c *Client) ListScripts() ([]string, error) {
	scriptsDir := filepath.Join(c.baseDir, "scripts")
	files, err := os.ReadDir(scriptsDir)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		result = append(result, file.Name())
	}

	return result, nil
}
