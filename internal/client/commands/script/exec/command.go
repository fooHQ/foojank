package exec

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/fzz"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "execute",
		ArgsUsage: "<script-name>",
		Usage:     "Execute a script locally",
		Action:    action,
		Aliases:   []string{"exec"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)
	client := codebase.New(*conf.Codebase)
	return execAction(logger, client)(ctx, c)
}

func execAction(logger *slog.Logger, client *codebase.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() < 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		scriptName := c.Args().First()
		scriptPath, err := client.GetScript(scriptName)
		if err != nil {
			if os.IsNotExist(err) {
				err := fmt.Errorf("script '%s' not found", scriptName)
				logger.ErrorContext(ctx, err.Error())
				return err
			}

			err := fmt.Errorf("cannot execute a script: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		pkgPath, err := buildTempPackage(scriptPath)
		if err != nil {
			err := fmt.Errorf("cannot build a script: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		binPath := filepath.Join(os.TempDir(), fmt.Sprintf("runscript-%s", nuid.Next()))
		output, err := client.BuildRunscript(ctx, binPath)
		if err != nil {
			err := fmt.Errorf("cannot build runscript: %w\n%s", err, output)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = execRunscript(ctx, binPath, pkgPath, c.Args().Tail()...)
		if err != nil {
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Codebase == nil {
		return errors.New("codebase not configured")
	}

	return nil
}

func buildTempPackage(src string) (string, error) {
	dst := filepath.Join(os.TempDir(), fmt.Sprintf("fj%s.fzz", nuid.Next()))
	err := fzz.Build(src, dst)
	if err != nil {
		return "", err
	}

	return dst, nil
}

func execRunscript(ctx context.Context, binPath, pkgPath string, args ...string) error {
	cmd := exec.CommandContext(ctx, binPath, append([]string{pkgPath}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
