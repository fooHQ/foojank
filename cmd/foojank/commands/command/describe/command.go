package describe

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	protodaemon "github.com/foohq/foojank/proto/daemon"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		ArgsUsage: "<name>",
		Usage:     "Describe a command",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Agent,
				Usage: "describe a command of the specified agent",
			},
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
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
				Name:  flags.Credential,
				Usage: "set credential",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
		ShellComplete:   actions.CompleteCommandName,
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

	ctx, err = actions.SetupServer(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)
	agentName, _ := conf.String(flags.Agent)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := daemon.New(srv)

	commandName := c.Args().First()

	var command protodaemon.AgentCommand
	{
		resp, err := client.GetCommand(ctx, protodaemon.GetCommandRequest{
			AgentRef: agentName,
			Name:     commandName,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get command: %v", err)
			return err
		}

		command = resp.Command
	}

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("DESCRIPTION").WithBold(),
		formatter.NewStringCell("FLAGS").WithBold(),
		formatter.NewStringCell("ENVIRONMENT").WithBold(),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell(commandName),
		formatter.NewStringCell(command.Description),
		formatter.NewStringSliceCell(formatFlags(command.Flags)).WithSeparator("\n"),
		formatter.NewStringSliceCell(formatEnvironment(command.Environment)).WithSeparator("\n"),
	})

	err := formatter.NewFormatter(
		format,
		formatter.WithNoColor(noColor),
		formatter.WithOrientation(formatter.OrientationHorizontal),
	).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func formatFlags(flags map[string]protodaemon.AgentCommandFlag) []string {
	names := make([]string, 0, len(flags))
	for name := range flags {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := make([]string, 0, len(names))
	for _, name := range names {
		flag := flags[name]
		line := "--" + name

		var extras []string
		if flag.Required {
			extras = append(extras, "required")
		}
		if flag.Default != "" {
			extras = append(extras, "default: "+flag.Default)
		}
		if len(extras) > 0 {
			line += " (" + strings.Join(extras, ", ") + ")"
		}
		if flag.Description != "" {
			line += ": " + flag.Description
		}
		lines = append(lines, line)
	}
	return lines
}

func formatEnvironment(env map[string]protodaemon.AgentCommandVar) []string {
	names := make([]string, 0, len(env))
	for name := range env {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := make([]string, 0, len(names))
	for _, name := range names {
		variable := env[name]
		line := name + " = " + variable.Value
		if variable.Description != "" {
			line += " (" + variable.Description + ")"
		}
		lines = append(lines, line)
	}
	return lines
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Agent,
		flags.ServerURL,
		flags.Credential,
	} {
		switch opt {
		case flags.Agent:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("agent not configured")
			}
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Credential:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("credential not configured")
			}
		}
	}
	return nil
}
