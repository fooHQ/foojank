package codebase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/fzz"
)

var ErrScriptNotFound = fmt.Errorf("script not found")

type Client struct {
	baseDir string
}

// TODO: check baseDir!

func New(path string) *Client {
	return &Client{
		baseDir: path,
	}
}

func (c *Client) BuildDir() string {
	return filepath.Join(c.baseDir, "build")
}

func (c *Client) BuildAgent(ctx context.Context, os, arch string, production bool) (string, string, error) {
	script := "build-agent-dev"
	if production {
		script = "build-agent-prod"
	}
	output := filepath.Join(c.BuildDir(), nuid.Next())
	if os == "windows" {
		output += ".exe"
	}
	result, err := c.devboxRun(ctx, script, map[string]string{
		"GOOS":   os,
		"GOARCH": arch,
		"OUTPUT": output,
	})
	if err != nil {
		return "", result, err
	}
	return output, result, nil
}

func (c *Client) WriteAgentConfig(b []byte) error {
	confFile := filepath.Join(c.baseDir, "internal", "vessel", "config", "config.go")
	err := os.WriteFile(confFile, b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) BuildRunscript(ctx context.Context) (string, string, error) {
	output := filepath.Join(c.BuildDir(), fmt.Sprintf("runscript-%s", nuid.Next()))
	result, err := c.devboxRun(ctx, "build-runscript", map[string]string{
		"OUTPUT": output,
	})
	if err != nil {
		return "", result, err
	}
	return output, result, nil
}

func (c *Client) WriteRunscriptConfig(b []byte) error {
	confFile := filepath.Join(c.baseDir, "internal", "runscript", "config", "config.go")
	err := os.WriteFile(confFile, b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetScript(name string) (string, error) {
	scriptsDir := filepath.Join(c.baseDir, "scripts", name)
	_, err := os.ReadDir(scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrScriptNotFound
		}
		return "", err
	}

	return scriptsDir, nil
}

func (c *Client) BuildScript(name string) (string, error) {
	scriptSrc, err := c.GetScript(name)
	if err != nil {
		return "", err
	}

	outputName := filepath.Join(c.BuildDir(), nuid.Next())
	err = fzz.Build(scriptSrc, outputName)
	if err != nil {
		return "", err
	}

	return outputName, nil
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

func (c *Client) ListModules() ([]string, error) {
	scriptsDir := filepath.Join(c.baseDir, "modules")
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
