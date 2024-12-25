package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/formatter"
	jsonformatter "github.com/foohq/foojank/internal/client/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/client/formatter/table"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagFormat = "format"
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
		},
		Action:  action,
		Aliases: []string{"ls"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Servers, conf.User.JWT, conf.User.KeySeed)
	if err != nil {
		err := fmt.Errorf("cannot connect to the server: %v", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %v", err)
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
				files, err := client.ListFiles(ctx, r)
				if err != nil {
					err := fmt.Errorf("cannot list contents of repository '%s': %v", r, err)
					logger.ErrorContext(ctx, err.Error())
					return err
				}

				table := formatter.NewTable([]string{
					"name",
					"size",
					"modified",
				})
				for _, file := range files {
					name := file.Name
					size := strconv.FormatUint(file.Size, 10)
					modified := file.Modified.String()
					table.AddRow([]string{
						name,
						size,
						modified,
					})
				}

				err = f.Write(os.Stdout, table)
				if err != nil {
					err := fmt.Errorf("cannot write formatted output: %v", err)
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
			name := repo.Name
			size := strconv.FormatUint(repo.Size, 10)
			description := repo.Description
			table.AddRow([]string{
				name,
				size,
				description,
			})
		}

		err = f.Write(os.Stdout, table)
		if err != nil {
			err := fmt.Errorf("cannot write formatted output: %v", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Servers == nil {
		return errors.New("servers not configured")
	}

	if conf.User == nil {
		return errors.New("user not configured")
	}

	return nil
}
