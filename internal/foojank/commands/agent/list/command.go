package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
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
	// TODO: rename to filter-name
	FlagServiceName      = "service-name"
	FlagTimeout          = "timeout"
	FlagFormat           = "format"
	FlagServer           = flags.Server
	FlagUserJWT          = flags.UserJWT
	FlagUserKey          = flags.UserKey
	FlagTLSCACertificate = flags.TLSCACertificate
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List active agents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagServiceName,
				Usage: "filter by service name",
			},
			&cli.DurationFlag{
				Name:  FlagTimeout,
				Usage: "set wait timeout",
				Value: 2 * time.Second,
			},
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
				err := fmt.Errorf("cannot write formatted output: %w", err)
				logger.ErrorContext(ctx, err.Error())
				return
			}
		}()

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := client.Discover(ctx, serviceName, outputCh)
		if err != nil {
			err := fmt.Errorf("discovery request failed: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		close(outputCh)
		wg.Wait()

		return nil
	}
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
