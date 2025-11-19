package list

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
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
				Name:  flags.ServerCertificate,
				Usage: "set server TLS certificate",
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
		Before:       before,
		Action:       action,
		Aliases:      []string{"ls"},
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

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	format, _ := conf.String(flags.Format)

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

	client := vessel.New(srv)

	results, err := client.Discover(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of agents: %v", err)
		return err
	}

	sortedResults := slices.SortedFunc(maps.Values(results), func(v1, v2 vessel.DiscoverResult) int {
		if v1.LastSeen.Before(v2.LastSeen) {
			return -1
		}
		if v1.LastSeen.After(v2.LastSeen) {
			return 1
		}
		return 0
	})

	err = formatOutput(os.Stdout, format, sortedResults)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func formatOutput(w io.Writer, format string, data []vessel.DiscoverResult) error {
	table := formatter.NewTable([]string{
		"agent_id",
		"userhost",
		"system",
		"address",
		"updated",
	})
	for _, service := range data {
		table.AddRow([]string{
			service.AgentID,
			fmt.Sprintf("%s@%s", service.Username, service.Hostname),
			service.System,
			service.Address,
			formatTime(service.LastSeen),
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
	if t.IsZero() {
		return "never"
	}

	now := time.Now()
	diff := now.Sub(t)

	// Handle future dates
	if diff < 0 {
		diff = -diff
		if diff < 24*time.Hour {
			if diff < time.Hour {
				return fmt.Sprintf("in %d minutes", int(diff.Minutes()))
			}
			return fmt.Sprintf("in %d hours", int(diff.Hours()))
		}
		return fmt.Sprintf("in %d days", int(diff.Hours()/24))
	}

	// Handle past dates
	if diff < 24*time.Hour {
		if diff < 2*time.Minute {
			return "now"
		}
		if diff < time.Hour {
			return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
		}
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	}

	return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
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
