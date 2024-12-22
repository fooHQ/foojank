package codebase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

func (c *Client) BuildAgent(ctx context.Context, os, arch, output string, production bool) (string, error) {
	scriptName := "build-agent-dev"
	if production {
		scriptName = "build-agent-prod"
	}
	return c.devboxRun(ctx, scriptName, map[string]string{
		"OUTPUT": output,
	})
}

func (c *Client) BuildRunscript(ctx context.Context, output string) (string, error) {
	return c.devboxRun(ctx, "build-runscript", map[string]string{
		"OUTPUT": output,
	})
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

func (c *Client) devboxRun(ctx context.Context, script string, env map[string]string) (string, error) {
	environ := os.Environ()
	for name, value := range env {
		environ = append(environ, fmt.Sprintf("%s=%s", name, value))
	}

	cmd := exec.CommandContext(ctx, "devbox", "run", script)
	cmd.Dir = c.baseDir
	cmd.Env = environ
	b, err := cmd.CombinedOutput()
	return string(b), err
}
