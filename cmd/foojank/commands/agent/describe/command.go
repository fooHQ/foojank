package describe

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		ArgsUsage: "<name>",
		Usage:     "Describe an agent",
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

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := daemon.New(srv)

	agentName := c.Args().First()

	agent, err := client.GetAgent(ctx, agentName)
	if err != nil {
		if errors.Is(err, daemon.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", agentName)
		}
		logger.ErrorContext(ctx, "Cannot get agent: %v", err)
		return err
	}

	var gatewayName string
	gateway, err := client.GetGateway(ctx, agent.GatewayID)
	if err != nil {
		if !errors.Is(err, daemon.ErrKeyNotFound) {
			logger.ErrorContext(ctx, "Cannot get gateway: %v", err)
			return err
		}
		gatewayName = agent.GatewayID
	} else {
		gatewayName = gateway.Name
	}

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("OS").WithBold(),
		formatter.NewStringCell("ARCH").WithBold(),
		formatter.NewStringCell("GATEWAY").WithBold(),
		formatter.NewStringCell("JWT").WithBold(),
		formatter.NewStringCell("KEY").WithBold(),
		formatter.NewStringCell("EXTRA").WithBold(),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell(agent.ID),
		formatter.NewStringCell(agent.Name),
		formatter.NewStringCell(agent.Config.OS),
		formatter.NewStringCell(agent.Config.Arch),
		formatter.NewStringCell(gatewayName),
		formatter.NewStringCell(agent.Config.UserJWT),
		formatter.NewStringCell(agent.Config.UserKey),
		formatter.NewStringMapCell(agent.Config.Extra).WithKeyValueSeparator(" = ").WithSeparator("\n"),
	})

	err = formatter.NewFormatter(
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
