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

	"github.com/foohq/ren/packager"
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

type BuildAgentOptions struct {
	OS         string
	Arch       string
	Production bool
	Tags       []string
	Config     BuildAgentConfig
}

type BuildAgentConfig struct {
	ID                           string
	Server                       string
	UserJWT                      string
	UserKey                      string
	CACertificate                string
	Stream                       string
	Consumer                     string
	InboxPrefix                  string
	ObjectStoreName              string
	SubjectApiWorkerStartT       string
	SubjectApiWorkerStopT        string
	SubjectApiWorkerWriteStdinT  string
	SubjectApiWorkerWriteStdoutT string
	SubjectApiWorkerStatusT      string
	SubjectApiConnInfoT          string
	SubjectApiReplyT             string
}

func (c *BuildAgentConfig) ToFlags() string {
	ldFlags := []string{
		"-X main.ID=%s",
		"-X main.Server=%s",
		"-X main.UserJWT=%s",
		"-X main.UserKey=%s",
		"-X main.CACertificate=%s",
		"-X main.Stream=%s",
		"-X main.Consumer=%s",
		"-X main.InboxPrefix=%s",
		"-X main.ObjectStoreName=%s",
		"-X main.SubjectApiWorkerStartT=%s",
		"-X main.SubjectApiWorkerStopT=%s",
		"-X main.SubjectApiWorkerWriteStdinT=%s",
		"-X main.SubjectApiWorkerWriteStdoutT=%s",
		"-X main.SubjectApiWorkerStatusT=%s",
		"-X main.SubjectApiConnInfoT=%s",
		"-X main.SubjectApiReplyT=%s",
	}
	flags := []string{
		"-race",
		fmt.Sprintf("-ldflags=\"%s\"", strings.Join(ldFlags, " ")),
	}
	return fmt.Sprintf(strings.Join(flags, " "),
		c.ID,
		c.Server,
		c.UserJWT,
		c.UserKey,
		c.CACertificate,
		c.Stream,
		c.Consumer,
		c.InboxPrefix,
		c.ObjectStoreName,
		c.SubjectApiWorkerStartT,
		c.SubjectApiWorkerStopT,
		c.SubjectApiWorkerWriteStdinT,
		c.SubjectApiWorkerWriteStdoutT,
		c.SubjectApiWorkerStatusT,
		c.SubjectApiConnInfoT,
		c.SubjectApiReplyT,
	)
}

func (c *Client) BuildAgent(ctx context.Context, opts BuildAgentOptions) (string, string, error) {
	err := c.baseDirExists()
	if err != nil {
		return "", "", err
	}

	script := "build-agent-dev"
	if opts.Production {
		script = "build-agent-prod"
	}

	output := filepath.Join(c.BuildDir(), nuid.Next())
	if opts.OS == "windows" {
		output += ".exe"
	}

	env := map[string]string{
		"GOOS":                              opts.OS,
		"GOARCH":                            opts.Arch,
		"OUTPUT":                            output,
		"TAGS":                              strings.Join(opts.Tags, " "),
		"ID":                                opts.Config.ID,
		"SERVER":                            opts.Config.Server,
		"USER_JWT":                          opts.Config.UserJWT,
		"USER_KEY":                          opts.Config.UserKey,
		"CA_CERTIFICATE":                    opts.Config.CACertificate,
		"STREAM":                            opts.Config.Stream,
		"CONSUMER":                          opts.Config.Consumer,
		"INBOX_PREFIX":                      opts.Config.InboxPrefix,
		"OBJECT_STORE_NAME":                 opts.Config.ObjectStoreName,
		"SUBJECT_API_WORKER_START_T":        opts.Config.SubjectApiWorkerStartT,
		"SUBJECT_API_WORKER_STOP_T":         opts.Config.SubjectApiWorkerStopT,
		"SUBJECT_API_WORKER_WRITE_STDIN_T":  opts.Config.SubjectApiWorkerWriteStdinT,
		"SUBJECT_API_WORKER_WRITE_STDOUT_T": opts.Config.SubjectApiWorkerWriteStdoutT,
		"SUBJECT_API_WORKER_STATUS_T":       opts.Config.SubjectApiWorkerStatusT,
		"SUBJECT_API_CONN_INFO_T":           opts.Config.SubjectApiConnInfoT,
		"SUBJECT_API_REPLY_T":               opts.Config.SubjectApiReplyT,
	}

	result, err := c.devboxRun(ctx, script, env)
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
	err = packager.Build(scriptSrc, outputName)
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
