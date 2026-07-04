package build

import (
	"context"
	"encoding/base64"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/builder"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "build",
		ArgsUsage: "<name>",
		Usage:     "Build an agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.SourceDir,
				Usage: "set path to a source code directory",
			},
			&cli.StringFlag{
				Name:    flags.Output,
				Usage:   "set path to an output file",
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set server account",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	outputName, _ := conf.String(flags.Output)
	sourceDir, _ := conf.String(flags.SourceDir)

	userJWT, userSeed, err := auth.ReadUser(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", accountName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	if c.Args().Len() < 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := daemon.New(srv)

	agentName := c.Args().First()

	agent, err := client.GetAgent(ctx, agentName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get agent %q: %v", agentName, err)
		return err
	}

	// Prepare builder environment.
	// IMPORTANT: make sure mandatory variables such as OS, Arch, etc. are copied as last to prevent being overwritten
	// by variables from .Extra.
	env := make(map[string]string)
	maps.Copy(env, agent.Config.Extra)
	maps.Copy(env, map[string]string{
		builder.OS:                agent.Config.OS,
		builder.Arch:              agent.Config.Arch,
		builder.Target:            createTargetPath(agent.Config.OS, outputName),
		builder.AgentID:           agent.ID,
		builder.AgentName:         agent.Name,
		builder.ServerURL:         agent.Config.ServerURL,
		builder.ServerCertificate: base64.StdEncoding.EncodeToString(agent.Config.ServerCertificate),
	})

	output, err := builder.Run(ctx, sourceDir, env)
	if err != nil {
		lines := strings.SplitSeq(output, "\n")
		for line := range lines {
			if line == "" {
				continue
			}
			logger.ErrorContext(ctx, "%s", line)
		}
		logger.ErrorContext(ctx, "Build failed: %v", err)
		// Return a generic error instead of "err".
		// Err can be of type exit.ExitError, which is apparently printed to stderr by cli.
		return errors.New("build failed")
	}
	defer func() {
		if err == nil {
			return
		}
		err := os.Remove(outputName)
		if err != nil {
			logger.WarnContext(ctx, "Cannot remove executable file %q: %v", outputName, err)
		}
	}()

	logger.InfoContext(ctx, "Agent %q has been built!", agentName)

	return nil
}

func createTargetPath(os, name string) string {
	if os == "windows" && filepath.Ext(name) != ".exe" {
		name += ".exe"
	}
	pwd, err := filepath.Abs(".")
	if err != nil {
		return name
	}
	return filepath.Join(pwd, name)
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.SourceDir,
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
		case flags.SourceDir:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("source directory not configured")
			}
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
