package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func Run(ctx context.Context, dir string, env map[string]string) (string, error) {
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
