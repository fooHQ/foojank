package codebase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nats-io/nuid"
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

type BuildAgentOptions struct {
	OS         string
	Arch       string
	Production bool
	Tags       []string
	Config     BuildAgentConfig
}

type BuildAgentConfig struct {
	AgentID               string
	ServerURL             string
	ServerCertificate     string
	UserJWT               string
	UserKey               string
	Stream                string
	Consumer              string
	InboxPrefix           string
	ObjectStore           string
	AwaitMessagesDuration time.Duration
	IdleDuration          time.Duration
	IdleJitter            time.Duration
}

func (c *Client) BuildAgent(ctx context.Context, opts BuildAgentOptions) (string, string, error) {
	err := c.baseDirExists()
	if err != nil {
		return "", "", err
	}

	output := filepath.Join(c.BuildDir(), nuid.Next())
	if opts.OS == "windows" {
		output += ".exe"
	}

	env := map[string]string{
		"GOOS":                       opts.OS,
		"GOARCH":                     opts.Arch,
		"OUTPUT":                     output,
		"TAGS":                       strings.Join(opts.Tags, " "),
		"FJ_AGENT_ID":                opts.Config.AgentID,
		"FJ_SERVER_URL":              opts.Config.ServerURL,
		"FJ_SERVER_CERTIFICATE":      opts.Config.ServerCertificate,
		"FJ_USER_JWT":                opts.Config.UserJWT,
		"FJ_USER_KEY":                opts.Config.UserKey,
		"FJ_STREAM":                  opts.Config.Stream,
		"FJ_CONSUMER":                opts.Config.Consumer,
		"FJ_INBOX_PREFIX":            opts.Config.InboxPrefix,
		"FJ_OBJECT_STORE":            opts.Config.ObjectStore,
		"FJ_AWAIT_MESSAGES_DURATION": opts.Config.AwaitMessagesDuration.String(),
		"FJ_IDLE_INTERVAL":           opts.Config.IdleDuration.String(),
		"FJ_IDLE_JITTER":             opts.Config.IdleJitter.String(),
	}

	result, err := c.devboxRun(ctx, "build", env)
	if err != nil {
		return "", result, err
	}

	return output, result, nil
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
