package list

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/agent"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
	jsonformatter "github.com/foohq/foojank/internal/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/formatter/table"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List jobs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Agent,
				Usage: "filter jobs by agent",
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
	agentName, _ := conf.String(flags.Agent)
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

	client := agent.New(srv)

	var jobs map[string]agent.Job
	if agentName != "" {
		var agentID string
		agentID, err = client.GetAgentID(ctx, agentName)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get a list of jobs: %v", err)
			return err
		}

		jobs, err = client.ListJobs(ctx, agentID)
	} else {
		jobs, err = client.ListAllJobs(ctx)
	}
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of jobs: %v", err)
		return err
	}

	data := make([]agent.Job, 0, len(jobs))
	for _, job := range jobs {
		data = append(data, job)
	}

	sort.SliceStable(data, func(i, j int) bool {
		return data[i].Updated.Before(data[j].Updated)
	})

	err = formatOutput(os.Stdout, format, data)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func formatOutput(w io.Writer, format string, data []agent.Job) error {
	table := formatter.NewTable([]string{
		"job_id",
		"agent",
		"command",
		"last_update",
		"status",
	})
	for _, job := range data {
		table.AddRow([]string{
			job.ID,
			job.AgentName,
			fmt.Sprintf("%s %s", job.Command, job.Args),
			formatTime(job.Updated),
			strings.ToUpper(job.Status),
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
