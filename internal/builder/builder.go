package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Options struct {
	OS                string
	Architecture      string
	Features          []string
	AgentID           string
	ServerURL         string
	ServerCertificate string
	UserJWT           string
	UserKey           string
	Stream            string
	Consumer          string
	InboxPrefix       string
	ObjectStore       string
	Custom            map[string]string
}

func Build(ctx context.Context, dir, target string, opts Options) (string, error) {
	if opts.OS == "windows" && !strings.HasSuffix(target, ".exe") {
		target += ".exe"
	}

	env := map[string]string{
		"OS":                    opts.OS,
		"ARCH":                  opts.Architecture,
		"TARGET":                target,
		"FEATURES":              strings.Join(opts.Features, ","),
		"FJ_AGENT_ID":           opts.AgentID,
		"FJ_SERVER_URL":         opts.ServerURL,
		"FJ_SERVER_CERTIFICATE": opts.ServerCertificate,
		"FJ_USER_JWT":           opts.UserJWT,
		"FJ_USER_KEY":           opts.UserKey,
		"FJ_STREAM":             opts.Stream,
		"FJ_CONSUMER":           opts.Consumer,
		"FJ_INBOX_PREFIX":       opts.InboxPrefix,
		"FJ_OBJECT_STORE":       opts.ObjectStore,
	}

	for k, v := range opts.Custom {
		env[k] = v
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
