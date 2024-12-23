package build

import (
	"context"
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
	"github.com/foohq/foojank/internal/client/log"
	"github.com/foohq/foojank/internal/config"
)

const (
	FlagOs          = "os"
	FlagArch        = "arch"
	FlagOutput      = "output"
	FlagDev         = "dev"
	FlagAgentServer = "agent-server"
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
			&cli.StringFlag{
				Name: FlagAgentServer,
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

		agentName := nuid.Next()
		if outputName == "" {
			wd, err := os.Getwd()
			if err != nil {
				err := fmt.Errorf("cannot build an agent: cannot determine current working directory")
				logger.Error(err.Error())
				return err
			}

			outputName = filepath.Join(wd, agentName)
		}

		if targetOs == "windows" && !strings.HasSuffix(outputName, ".exe") {
			outputName += ".exe"
		}

		servers := conf.Servers
		if c.IsSet(FlagAgentServer) {
			servers = []string{c.String(FlagAgentServer)}
		}
		if servers == nil {
			err := fmt.Errorf("cannot build an agent: no server configured")
			logger.Error(err.Error())
			return err
		}

		accountClaims, err := jwt.DecodeAccountClaims(conf.Account.JWT)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot decode account JWT: %v", err)
			logger.Error(err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		user, err := config.NewUserAgent(agentName, accountPubKey, []byte(conf.Account.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		agentConf := config.Config{
			Servers: servers,
			User: &config.Entity{
				JWT:     user.JWT,
				KeySeed: user.KeySeed,
			},
			Service: &config.Service{
				Name:    agentName,
				Version: foojank.Version(),
			},
		}

		template := NewTemplate()
		confOutput, err := template.Render(agentConf)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		confFile := filepath.Join(*conf.Codebase, "internal", "vessel", "config", "config.go")
		err = os.WriteFile(confFile, confOutput, 0600)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot write agent configuration to file '%s': %v", confFile, err)
			logger.Error(err.Error())
			return err
		}

		output, err := client.BuildAgent(ctx, targetOs, targetArch, outputName, !devBuild)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: %v\n%s", err, output)
			logger.Error(err.Error())
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, filepath.Base(outputName))

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Codebase == nil {
		return fmt.Errorf("codebase not configured")
	}

	if conf.Servers == nil {
		return fmt.Errorf("servers not configured")
	}

	if conf.Account == nil {
		return fmt.Errorf("account not configured")
	}

	return nil
}
