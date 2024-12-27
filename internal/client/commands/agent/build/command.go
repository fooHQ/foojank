package build

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagOs            = "os"
	FlagArch          = "arch"
	FlagOutput        = "output"
	FlagDev           = "dev"
	FlagWithServer    = "with-server"
	FlagWithoutModule = "without-module"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagOs,
				Usage: "set build operating system",
				Value: runtime.GOOS,
			},
			&cli.StringFlag{
				Name:  FlagArch,
				Usage: "set build architecture",
				Value: runtime.GOARCH,
			},
			&cli.BoolFlag{
				Name:  FlagDev,
				Usage: "enable development mode",
				Value: false,
			},
			&cli.StringFlag{
				Name:    FlagOutput,
				Usage:   "set output file",
				Aliases: []string{"o"},
			},
			&cli.StringSliceFlag{
				Name:  FlagWithServer,
				Usage: "set agent's server",
			},
			&cli.StringSliceFlag{
				Name:  FlagWithoutModule,
				Usage: "disable compilation of a module",
			},
		},
		Action: action,
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
	return buildAction(logger, conf, client)(ctx, c)
}

func buildAction(logger *slog.Logger, conf *config.Config, client *codebase.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		outputName := c.String(FlagOutput)
		targetOs := c.String(FlagOs)
		targetArch := c.String(FlagArch)
		devBuild := c.Bool(FlagDev)
		isAgentServer := c.IsSet(FlagWithServer)
		agentServer := c.StringSlice(FlagWithServer)
		disabledModules := c.StringSlice(FlagWithoutModule)

		agentName := nuid.Next()
		if outputName == "" {
			wd, err := os.Getwd()
			if err != nil {
				err := fmt.Errorf("cannot build an agent: cannot determine current working directory")
				logger.ErrorContext(ctx, err.Error())
				return err
			}

			outputName = filepath.Join(wd, agentName)
		}

		if targetOs == "windows" && !strings.HasSuffix(outputName, ".exe") {
			outputName += ".exe"
		}

		servers := conf.Servers
		if isAgentServer {
			servers = agentServer
		}
		if servers == nil {
			err := errors.New("cannot build an agent: no server configured")
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		modules, err := client.ListModules()
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot get a list of modules: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		modules = configureModules(modules, disabledModules)

		accountClaims, err := jwt.DecodeAccountClaims(conf.Account.JWT)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot decode account JWT: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		user, err := config.NewUserAgent(agentName, accountPubKey, []byte(conf.Account.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		agentConf := templateData{
			Servers: servers,
			User: templateDataUser{
				JWT:     user.JWT,
				KeySeed: user.KeySeed,
			},
			Service: templateDataService{
				Name:    agentName,
				Version: foojank.Version(),
			},
			Modules: modules,
		}
		confOutput, err := RenderTemplate(templateString, agentConf)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = client.WriteAgentConfig(confOutput)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot write agent configuration to a file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		binPath, output, err := client.BuildAgent(ctx, targetOs, targetArch, !devBuild)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: %w\n%s", err, output)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = os.Rename(binPath, outputName)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot rename the executable file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, filepath.Base(outputName))

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

	if conf.Codebase == nil {
		return errors.New("codebase not configured")
	}

	if conf.Servers == nil {
		return errors.New("servers not configured")
	}

	if conf.Account == nil {
		return errors.New("account not configured")
	}

	return nil
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
