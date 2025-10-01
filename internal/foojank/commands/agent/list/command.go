package list

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagFormat           = "format"
	FlagServer           = flags.Server
	FlagUserJWT          = flags.UserJWT
	FlagUserKey          = flags.UserKey
	FlagTLSCACertificate = flags.TLSCACertificate
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List agents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagFormat,
				Usage: "set output format",
				Value: "table",
			},
			&cli.StringSliceFlag{
				Name:  FlagServer,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  FlagUserJWT,
				Usage: "set user JWT token",
			},
			&cli.StringFlag{
				Name:  FlagUserKey,
				Usage: "set user secret key",
			},
			&cli.StringFlag{
				Name:  FlagTLSCACertificate,
				Usage: "set TLS CA certificate",
			},
		},
		Action:       action,
		Aliases:      []string{"ls"},
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	srv, err := server.New(conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey, *conf.Client.TLSCACertificate)
	if err != nil {
		log.Error(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	client := vessel.New(srv)

	format := c.String(FlagFormat)

	results, err := client.Discover(ctx)
	if err != nil {
		log.Error(ctx, "Cannot get a list of agents: %v", err)
		return err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].LastSeen.Before(results[j].LastSeen)
	})

	err = formatOutput(os.Stdout, format, results)
	if err != nil {
		log.Error(ctx, "Cannot write formatted output: %v", err)
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
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	if conf.Client == nil {
		return errors.New("client configuration is missing")
	}

	if len(conf.Client.Server) == 0 {
		return errors.New("server not configured")
	}

	if conf.Client.UserJWT == nil {
		return errors.New("user jwt not configured")
	}

	if conf.Client.UserKey == nil {
		return errors.New("user key not configured")
	}

	if conf.Client.TLSCACertificate == nil {
		return errors.New("tls ca certificate not configured")
	}

	return nil
}
