package list

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/formatter"
	jsonformatter "github.com/foohq/foojank/internal/client/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/client/formatter/table"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagServiceName = "service-name"
	FlagTimeout     = "timeout"
	FlagFormat      = "format"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List active agents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagServiceName,
				Usage: "filter by service name",
				Value: "",
			},
			&cli.DurationFlag{
				Name:  FlagTimeout,
				Usage: "set how long to wait for response",
				Value: 2 * time.Second,
			},
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
		logger.Error(err.Error())
		return err
	}

	client := vessel.New(nc)
	return listAction(logger, client)(ctx, c)
}

func listAction(logger *slog.Logger, client *vessel.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		serviceName := c.String(FlagServiceName)
		timeout := c.Duration(FlagTimeout)
		format := c.String(FlagFormat)

		outputCh := make(chan vessel.Service)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			table := formatter.NewTable([]string{
				"id",
				"user",
				"hostname",
				"system",
				"ip_address",
			})
			for service := range outputCh {
				logger.Debug("found a service", "service", service)

				id := service.ID.String()
				ip, ipOk := service.Metadata["ip_address"]
				user, userOk := service.Metadata["user"]
				hostname, hostnameOk := service.Metadata["hostname"]
				osName, osNameOk := service.Metadata["os"]

				// This condition filters out workers and incompatible services.
				// Detection should work even if worker was not able to determine
				// a value, in such case the value will be an empty string.
				if !ipOk || !userOk || !hostnameOk || !osNameOk {
					continue
				}

				table.AddRow([]string{
					id,
					user,
					hostname,
					osName,
					ip,
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
				err := fmt.Errorf("unknown output format '%s', using the default format instead", format)
				logger.Warn(err.Error())
			}

			err := f.Write(os.Stdout, table)
			if err != nil {
				err := fmt.Errorf("cannot write formatted output: %v", err)
				logger.Error(err.Error())
				return
			}
		}()

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := client.Discover(ctx, serviceName, outputCh)
		if err != nil {
			err := fmt.Errorf("discovery request failed: %v", err)
			logger.Error(err.Error())
			return err
		}

		close(outputCh)
		wg.Wait()

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Servers == nil {
		return fmt.Errorf("servers not configured")
	}

	if conf.User == nil {
		return fmt.Errorf("user not configured")
	}

	return nil
}
