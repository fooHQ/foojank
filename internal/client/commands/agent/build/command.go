package build

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "project-dir",
				// TODO: use something else!
				Value: "/tmp/foojank",
			},
		},
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	return buildAction(logger, conf)(ctx, c)
}

func buildAction(logger *slog.Logger, conf *config.Config) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		srcDir := c.String("project-dir")
		// TODO: configurable OS

		username := nuid.Next()
		account := conf.Account
		if account == nil {
			err := fmt.Errorf("cannot generate agent configuration: no account found")
			logger.Error(err.Error())
			return err
		}

		user, err := config.NewUserAgent(username, account.PublicKey, []byte(account.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot generate agent configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		agentConf := config.Config{
			Servers: conf.Servers,
			User: &config.Entity{
				JWT:       user.JWT,
				PublicKey: user.PublicKey,
				KeySeed:   user.KeySeed,
			},
			Service: &config.Service{
				Name:    username,
				Version: foojank.Version(),
			},
		}

		template := NewTemplate()
		output, err := template.Render(agentConf)
		if err != nil {
			err := fmt.Errorf("cannot generate agent configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		confFile := filepath.Join(srcDir, "internal", "vessel", "config", "config.go")
		err = os.WriteFile(confFile, output, 0600)
		if err != nil {
			err := fmt.Errorf("cannot write agent configuration to file '%s': %v", confFile, err)
			logger.Error(err.Error())
			return err
		}

		err = exec.CommandContext(ctx, "devbox", "run", "build-agent-prod").Run()
		if err != nil {
			err := fmt.Errorf("cannot build agent: %v", err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}
