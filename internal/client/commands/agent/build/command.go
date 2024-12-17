package build

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/flags"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "os",
			},
			&cli.StringFlag{
				Name: "arch",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
			},
			// TODO: configurable servers (for the agent)!
			&cli.StringFlag{
				Name:  flags.Codebase,
				Usage: "path to directory with foojank codebase",
				Value: flags.DefaultCodebase(),
			},
			&cli.StringFlag{
				Name:  flags.AccountJWT,
				Usage: "account JWT token",
			},
			&cli.StringFlag{
				Name:  flags.AccountSigningKey,
				Usage: "account signing key",
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
		outputName := c.String("output")
		targetOs := c.String("os")
		targetArch := c.String("arch")

		// TODO: move to validation function
		if conf.Codebase == nil {
			err := fmt.Errorf("cannot build an agent: codebase not configured")
			logger.Error(err.Error())
			return err
		}

		if outputName == "" {
			outputName = nuid.Next()
		}

		if targetOs == "windows" && !strings.HasSuffix(outputName, ".exe") {
			outputName += ".exe"
		}

		username := nuid.Next()
		account := conf.Account
		if account == nil {
			err := fmt.Errorf("cannot build an agent: no account found")
			logger.Error(err.Error())
			return err
		}

		accountClaims, err := jwt.DecodeAccountClaims(account.JWT)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot decode account JWT: %v", err)
			logger.Error(err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		user, err := config.NewUserAgent(username, accountPubKey, []byte(account.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		agentConf := config.Config{
			Servers: conf.Servers,
			User: &config.Entity{
				JWT:     user.JWT,
				KeySeed: user.KeySeed,
			},
			Service: &config.Service{
				Name:    username,
				Version: foojank.Version(),
			},
		}

		template := NewTemplate()
		output, err := template.Render(agentConf)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		confFile := filepath.Join(*conf.Codebase, "internal", "vessel", "config", "config.go")
		err = os.WriteFile(confFile, output, 0600)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot write agent configuration to file '%s': %v", confFile, err)
			logger.Error(err.Error())
			return err
		}

		env := os.Environ()
		if targetOs != "" {
			env = append(env, "GOOS="+targetOs)
		}
		if targetArch != "" {
			env = append(env, "GOARCH="+targetArch)
		}
		env = append(env, "OUTPUT="+outputName)

		cmd := exec.CommandContext(ctx, "devbox", "run", "build-agent-prod")
		cmd.Dir = *conf.Codebase
		cmd.Env = env
		b, err := cmd.CombinedOutput()
		if err != nil {
			err := fmt.Errorf("cannot build an agent: %v\n%s", err, string(b))
			logger.Error(err.Error())
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, outputName)

		return nil
	}
}
