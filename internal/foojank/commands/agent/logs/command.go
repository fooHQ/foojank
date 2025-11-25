package logs

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "logs",
		ArgsUsage: "<agent-id>",
		Usage:     "Display all messages in agent's stream",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  config.Format,
				Usage: "set output format",
			},
			&cli.StringFlag{
				Name:  config.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  config.ServerCertificate,
				Usage: "set server TLS certificate",
			},
			&cli.StringFlag{
				Name:  config.Account,
				Usage: "set server account",
			},
			&cli.StringFlag{
				Name:  config.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:       before,
		Action:       action,
		OnUsageError: actions.UsageError,
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

	serverURL, _ := conf.String(config.ServerURL)
	serverCert, _ := conf.String(config.ServerCertificate)
	accountName, _ := conf.String(config.Account)
	format, _ := conf.String(config.Format)

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

	client := vessel.New(srv)

	agentID := c.Args().First()

	msgs, err := client.ListMessages(ctx, agentID, nil)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of messages: %v", err)
		return err
	}

	err = formatOutput(os.Stdout, format, msgs)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func formatOutput(w io.Writer, format string, data []vessel.Message) error {
	table := formatter.NewTable([]string{
		"message_id",
		"subject",
		"received",
	})
	for _, msg := range data {
		table.AddRow([]string{
			msg.ID,
			msg.Subject,
			formatTime(msg.Received),
		})
	}

	var f formatter.Formatter
	switch format {
	case "json":
		f = jsonformatter.New()
	case "table":
		f = tableformatter.New()
	default:
		f = tableformatter.New()
	}

	err := f.Write(w, table)
	if err != nil {
		return err
	}

	return nil
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		config.ServerURL,
		config.Account,
	} {
		switch opt {
		case config.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case config.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
