package exec

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagWithoutModule = "without-module"
	FlagDataDir       = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "execute",
		ArgsUsage: "<script-name>",
		Usage:     "Execute a script locally",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  FlagWithoutModule,
				Usage: "disable compilation of a module",
			},
			&cli.StringFlag{
				Name:  FlagDataDir,
				Usage: "set path to a data directory",
			},
		},
		Action:  action,
		Aliases: []string{"exec"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: cannot parse configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)
	client := codebase.New(*conf.DataDir)
	return execAction(logger, client)(ctx, c)
}

func execAction(logger *slog.Logger, client *codebase.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() < 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		disabledModules := c.StringSlice(FlagWithoutModule)
		scriptName := c.Args().First()

		pkgPath, err := client.BuildScript(scriptName)
		if err != nil {
			err := fmt.Errorf("cannot build script '%s': %w", scriptName, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		modules, err := client.ListModules()
		if err != nil {
			err := fmt.Errorf("cannot get a list of modules: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		modules = configureModules(modules, disabledModules)

		runscriptConf := templateData{
			Modules: modules,
		}
		confOutput, err := RenderTemplate(templateString, runscriptConf)
		if err != nil {
			err := fmt.Errorf("cannot generate runscript configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = client.WriteRunscriptConfig(confOutput)
		if err != nil {
			err := fmt.Errorf("cannot write runscript configuration to a file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		binPath, result, err := client.BuildRunscript(ctx)
		if err != nil {
			err := fmt.Errorf("cannot build runscript: %w\n%s", err, result)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = execRunscript(ctx, binPath, pkgPath, c.Args().Slice()...)
		if err != nil {
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	if conf.DataDir == nil {
		return errors.New("codebase not configured")
	}

	return nil
}

func execRunscript(ctx context.Context, binPath, pkgPath string, args ...string) error {
	cmd := exec.CommandContext(ctx, binPath, append([]string{pkgPath}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func configureModules(enabled, disabled []string) []string {
	var result []string
	for _, e := range enabled {
		found := false
		for _, d := range disabled {
			if e == d {
				found = true
			}
		}
		if !found {
			result = append(result, e)
		}
	}
	return result
}
