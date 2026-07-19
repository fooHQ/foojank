package list

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List agents",
		Flags: []cli.Flag{
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
		Aliases:         []string{"ls"},
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

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

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

	client := daemon.New(srv)

	agents, err := client.ListAgents(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of agents: %v", err)
		return err
	}

	agentHosts, err := client.ListAgentHosts(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of agent hosts: %v", err)
		return err
	}

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("USERHOST").WithBold(),
		formatter.NewStringCell("ADDRESS").WithBold(),
		formatter.NewStringCell("PLATFORM").WithBold(),
		formatter.NewStringCell("LAST SEEN").WithBold(),
	})

	for _, agent := range agents {
		var host daemon.AgentHostDirectoryEntry
		for i := range agentHosts {
			if agent.ID == agentHosts[i].AgentID {
				host = agentHosts[i]
			}
		}

		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(agent.Name),
			formatter.NewStringCell(formatUserHost(host.Username, host.Hostname)),
			formatter.NewStringCell(host.Address),
			formatter.NewStringCell(formatPlatform(agent.Config.OS, agent.Config.Arch)),
			formatter.NewTimeCell(host.LastUpdate).WithFormat("relative").WithEmptyValue("never"),
		})
	}

	err = formatter.NewFormatter(
		format,
		formatter.WithNoColor(noColor),
		formatter.WithSortByColumn(4, 0),
	).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func formatUserHost(user, host string) string {
	if user == "" && host == "" {
		return ""
	}
	return strings.Join([]string{user, host}, "@")
}

func formatPlatform(agentOS, agentArch string) string {
	if agentOS == "" && agentArch == "" {
		return ""
	}
	return strings.Join([]string{agentOS, agentArch}, "/")
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
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
