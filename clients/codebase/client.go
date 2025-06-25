package codebase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/fzz"
)

var (
	ErrScriptNotFound = errors.New("script not found")
)

type Client struct {
	baseDir string
}

func New(path string) *Client {
	return &Client{
		baseDir: filepath.Join(path, "src"),
	}
}

func (c *Client) BuildDir() string {
	return filepath.Join(c.baseDir, "build")
}

func (c *Client) ModulesDir() string {
	return filepath.Join(c.baseDir, "internal/engine/modules")
}

func (c *Client) ScriptsDir() string {
	return filepath.Join(c.baseDir, "scripts")
}

func (c *Client) VesselConfigFile() string {
	return filepath.Join(c.baseDir, "internal", "vessel", "config", "config.go")
}

func (c *Client) RunscriptConfigFile() string {
	return filepath.Join(c.baseDir, "internal", "runscript", "config", "config.go")
}

func (c *Client) BuildAgent(ctx context.Context, os, arch string, production bool) (string, string, error) {
	err := c.baseDirExists()
	if err != nil {
		return "", "", err
	}

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
	err := c.baseDirExists()
	if err != nil {
		return nil
	}

	err = os.WriteFile(c.VesselConfigFile(), b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) BuildRunscript(ctx context.Context, tags []string) (string, string, error) {
	err := c.baseDirExists()
	if err != nil {
		return "", "", err
	}

	output := filepath.Join(c.BuildDir(), fmt.Sprintf("runscript-%s", nuid.Next()))
	result, err := c.devboxRun(ctx, "build-runscript", map[string]string{
		"OUTPUT": output,
		"TAGS":   strings.Join(tags, " "),
	})
	if err != nil {
		return "", result, err
	}
	return output, result, nil
}

func (c *Client) WriteRunscriptConfig(b []byte) error {
	err := c.baseDirExists()
	if err != nil {
		return err
	}

	err = os.WriteFile(c.RunscriptConfigFile(), b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetScript(name string) (string, error) {
	err := c.baseDirExists()
	if err != nil {
		return "", err
	}

	scriptDir := filepath.Join(c.ScriptsDir(), name)
	_, err = os.ReadDir(scriptDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrScriptNotFound
		}
		return "", err
	}

	return scriptDir, nil
}

func (c *Client) BuildScript(name string) (string, error) {
	err := c.baseDirExists()
	if err != nil {
		return "", err
	}

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
	err := c.baseDirExists()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(c.ScriptsDir())
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
	err := c.baseDirExists()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(c.ModulesDir())
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

func (c *Client) baseDirExists() error {
	_, err := os.Stat(c.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("base directory '%s' does not exist", c.baseDir)
		}
	}
	return nil
}
