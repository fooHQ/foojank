// Package supervisor runs a go-plugin client and starts a new process when the
// plugin exits unexpectedly.
package supervisor

import (
	"context"
	"errors"
	"os/exec"
	"slices"
	"time"

	"github.com/hashicorp/go-plugin"
)

type Plugin struct {
	Name string

	Client plugin.ClientConfig

	// Started runs after every successful dispense, including each restart.
	// Returning an error kills this process and does not start it again.
	Started func(dispensed any) error

	// Stopped runs before the process is killed or after the process died.
	// err is set when the plugin process has already exited. Returning an
	// error stops this plugin instead of restarting it.
	Stopped func(dispensed any, err error) error

	// RestartDelay is how long to wait before launching again after the process
	// exits. The zero value waits one second.
	RestartDelay time.Duration
}

// Run launches the plugin and blocks until ctx is cancelled. An unexpected
// exit starts a new process. An error from starting the client, dispensing,
// or Started is returned and is not retried. Client.Cmd must be set.
func Run(ctx context.Context, cfg Plugin) error {
	delay := cfg.RestartDelay
	if delay <= 0 {
		delay = time.Second
	}

	for {
		err := run(ctx, cfg)
		if err != nil {
			return err
		}

		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func run(ctx context.Context, cfg Plugin) error {
	clientCfg := cfg.Client
	clientCfg.Cmd = cloneCmd(clientCfg.Cmd)

	c := plugin.NewClient(&clientCfg)
	defer c.Kill()

	proto, err := c.Client()
	if err != nil {
		return err
	}

	dispensed, err := proto.Dispense(cfg.Name)
	if err != nil {
		return err
	}

	if cfg.Started != nil {
		err := cfg.Started(dispensed)
		if err != nil {
			return err
		}
	}

	// Observe if client hasn't exited.
loop:
	for {
		if c.Exited() {
			err = errors.New("plugin exited")
			break
		}
		select {
		case <-time.After(500 * time.Millisecond):
			continue
		case <-ctx.Done():
			break loop
		}
	}

	if cfg.Stopped != nil {
		err := cfg.Stopped(dispensed, err)
		if err != nil {
			return err
		}
	}

	return nil
}

func cloneCmd(in *exec.Cmd) *exec.Cmd {
	out := &exec.Cmd{
		Path:        in.Path,
		Args:        slices.Clone(in.Args),
		Env:         slices.Clone(in.Env),
		Dir:         in.Dir,
		Stdin:       in.Stdin,
		Stdout:      in.Stdout,
		Stderr:      in.Stderr,
		ExtraFiles:  slices.Clone(in.ExtraFiles),
		SysProcAttr: in.SysProcAttr,
		Cancel:      in.Cancel,
		WaitDelay:   in.WaitDelay,
	}
	return out
}
