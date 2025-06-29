package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
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
		Name:      "list",
		ArgsUsage: "[repository]",
		Usage:     "List repositories or their contents",
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

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey, *conf.Client.TLSCACertificate)
	if err != nil {
		err := fmt.Errorf("cannot connect to the server: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	client := repository.New(js)
	return listAction(logger, client)(ctx, c)
}

func listAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		format := c.String(FlagFormat)

		var f formatter.Formatter
		switch format {
		case "json":
			f = jsonformatter.New()
		case "table":
			f = tableformatter.New()
		default:
			f = tableformatter.New()
			err := fmt.Errorf("unknown output format '%s', using the default format instead", format)
			logger.Warn(err.Error())
		}

		if c.Args().Len() > 0 {
			for _, r := range c.Args().Slice() {
				files, err := listDirectory(ctx, client, r, "/")
				if err != nil {
					err := fmt.Errorf("cannot list contents of repository '%s': %w", r, err)
					logger.ErrorContext(ctx, err.Error())
					return err
				}

				table := formatter.NewTable([]string{
					"name",
					"size",
					"modified",
				})
				for _, file := range files {
					name := file.Name()
					info, err := file.Info()
					if err != nil {
						err := fmt.Errorf("cannot get information about file '%s': %w", name, err)
						logger.ErrorContext(ctx, err.Error())
						return err
					}

					size := formatBytes(uint64(info.Size()))
					modified := formatTime(info.ModTime())
					table.AddRow([]string{
						name,
						size,
						modified,
					})
				}

				err = f.Write(os.Stdout, table)
				if err != nil {
					err := fmt.Errorf("cannot write formatted output: %w", err)
					logger.ErrorContext(ctx, err.Error())
					return err
				}
			}
			return nil
		}

		repos, err := client.List(ctx)
		if err != nil {
			return err
		}

		table := formatter.NewTable([]string{
			"name",
			"size",
			"description",
		})
		for _, repo := range repos {
			name := repo.Name()
			size := formatBytes(repo.Size())
			description := repo.Description()
			table.AddRow([]string{
				name,
				size,
				description,
			})
		}

		err = f.Write(os.Stdout, table)
		if err != nil {
			err := fmt.Errorf("cannot write formatted output: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		return nil
	}
}

func listDirectory(ctx context.Context, client *repository.Client, name, dir string) ([]risoros.DirEntry, error) {
	repo, err := client.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	defer repo.Close()

	err = repo.Wait(ctx)
	if err != nil {
		return nil, err
	}

	files, err := repo.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func formatBytes(size uint64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota) // 1 << 10 = 1024
		MB
		GB
		TB
	)

	var unit string
	var value float64

	switch {
	case size >= TB:
		value = float64(size) / TB
		unit = "TB"
	case size >= GB:
		value = float64(size) / GB
		unit = "GB"
	case size >= MB:
		value = float64(size) / MB
		unit = "MB"
	case size >= KB:
		value = float64(size) / KB
		unit = "kB"
	default:
		value = float64(size)
		unit = "B"
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
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
