package exec

import (
	"context"
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
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	if conf.Codebase == nil {
		err := fmt.Errorf("cannot execute a script: codebase not configured")
		logger.Error(err.Error())
		return err
	}

	client := codebase.New(*conf.Codebase)
	return execAction(logger, client)(ctx, c)
}

func execAction(logger *slog.Logger, client *codebase.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		scriptName := c.Args().First()
		scriptPath, err := client.GetScript(scriptName)
		if err != nil {
			if os.IsNotExist(err) {
				err := fmt.Errorf("script '%s' not found", scriptName)
				logger.Error(err.Error())
				return err
			}

			err := fmt.Errorf("cannot execute a script: %v", err)
			logger.Error(err.Error())
			return err
		}

		pkgPath, err := buildTempPackage(scriptPath)
		if err != nil {
			err := fmt.Errorf("cannot build a script: %v", err)
			logger.Error(err.Error())
			return err
		}

		runscriptBin := filepath.Join(os.TempDir(), fmt.Sprintf("runscript-%s", nuid.Next()))
		output, err := client.BuildRunscript(ctx, runscriptBin)
		if err != nil {
			err := fmt.Errorf("cannot build runscript: %v\n%s", err, output)
			logger.Error(err.Error())
			return err
		}

		cmd := exec.CommandContext(ctx, runscriptBin, pkgPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}

func buildTempPackage(src string) (string, error) {
	dst := filepath.Join(os.TempDir(), fmt.Sprintf("fj%s.fzz", nuid.Next()))
	err := fzz.Build(src, dst)
	if err != nil {
		return "", err
	}

	return dst, nil
}
