package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Options struct {
	OS       string
	Arch     string
	Features []string
	Config   Config
}

type Config struct {
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

func Build(ctx context.Context, dir, target string, opts Options) (string, error) {
	if opts.OS == "windows" && !strings.HasSuffix(target, ".exe") {
		target += ".exe"
	}

	env := map[string]string{
		"OS":                         opts.OS,
		"ARCH":                       opts.Arch,
		"TARGET":                     target,
		"FEATURES":                   strings.Join(opts.Features, ","),
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
		"FJ_IDLE_DURATION":           opts.Config.IdleDuration.String(),
		"FJ_IDLE_JITTER":             opts.Config.IdleJitter.String(),
	}

	result, err := devboxRun(ctx, dir, "build", env)
	if err != nil {
		return result, err
	}

	return result, nil
}

func devboxRun(ctx context.Context, dir, script string, env map[string]string) (string, error) {
	environ := os.Environ()
	for name, value := range env {
		environ = append(environ, fmt.Sprintf("%s=%s", name, value))
	}

	cmd := exec.CommandContext(ctx, "devbox", "run", script)
	cmd.Dir = dir
	cmd.Env = environ
	b, err := cmd.CombinedOutput()
	return string(b), err
}
