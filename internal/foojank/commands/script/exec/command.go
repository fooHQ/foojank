package exec

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/ren/modules"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagScript        = flags.Script
	FlagWithoutModule = "without-module"
	FlagDataDir       = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "execute",
		Usage: "Execute a script locally",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    FlagScript,
				Usage:   "script to execute",
				Aliases: []string{"s"},
			},
			&cli.StringSliceFlag{
				Name:  FlagWithoutModule,
				Usage: "disable compilation of a module",
			},
			&cli.StringFlag{
				Name:  FlagDataDir,
				Usage: "set path to a data directory",
			},
		},
		Action:       action,
		Aliases:      []string{"exec"},
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
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
		// Script arguments should include the name of the script as well.
		var scriptArgs []string
		var scriptName string
		if c.IsSet(FlagScript) {
			scriptArgs = strings.Fields(c.String(FlagScript))
			if len(scriptArgs) != 0 {
				scriptName = scriptArgs[0]
			}
		}

		disabledMods := c.StringSlice(FlagWithoutModule)

		pkgPath, err := client.BuildScript(scriptName)
		if err != nil {
			err := fmt.Errorf("cannot build script '%s': %w", scriptName, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}
		defer os.Remove(pkgPath)

		runscriptConf := templateData{}
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

		buildTags, err := configureBuildTags(modules.Modules(), disabledMods)
		if err != nil {
			err := fmt.Errorf("cannot configure modules: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		binPath, result, err := client.BuildRunscript(ctx, buildTags)
		if err != nil {
			err := fmt.Errorf("cannot build runscript: %w\n%s", err, result)
			logger.ErrorContext(ctx, err.Error())
			return err
		}
		defer os.Remove(binPath)

		err = execRunscript(ctx, binPath, pkgPath, scriptArgs)
		if err != nil && !errors.Is(err, context.Canceled) {
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

func execRunscript(ctx context.Context, binPath, pkgPath string, args []string) error {
	cmd := exec.CommandContext(ctx, binPath, pkgPath)
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	return cmd.Run()
}

func moduleExists(mods []string, name string) bool {
	for _, m := range mods {
		if m == name {
			return true
		}
	}
	return false
}

func configureBuildTags(mods, disabledMods []string) ([]string, error) {
	// Verify that disabled modules exist, otherwise throw an error.
	for _, m := range disabledMods {
		if !moduleExists(mods, m) {
			err := fmt.Errorf("module '%s' does not exist", m)
			return nil, err
		}
	}

	var buildTags []string
	for _, m := range mods {
		if moduleExists(disabledMods, m) {
			continue
		}
		buildTags = append(buildTags, modules.BuildTag(m))
	}
	return buildTags, nil
}
