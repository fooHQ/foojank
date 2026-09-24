package create

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/authdir"
	protodaemon "github.com/foohq/foojank/proto/daemon"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a gateway",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set gateway name",
			},
			&cli.StringFlag{
				Name:  flags.Description,
				Usage: "set gateway description",
			},
			&cli.StringSliceFlag{
				Name:  flags.Variable,
				Usage: "set configuration variable (format: key=value)",
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
				Name:  flags.Credential,
				Usage: "set credential",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
		ShellComplete:   actions.CompleteFlags,
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

func action(ctx context.Context, _ *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	credsName, _ := conf.String(flags.Credential)
	gatewayName, _ := conf.String(flags.Name)
	gatewayDesc, _ := conf.String(flags.Description)
	setVars, _ := conf.StringSlice(flags.Variable)

	userJWT, userSeed, err := authdir.ReadUser(credsName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", credsName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	client := daemon.New(srv)

	_, err = client.RequestCreateGateway(ctx, protodaemon.CreateGatewayRequest{
		Name:        gatewayName,
		Description: gatewayDesc,
		Config: protodaemon.GatewayConfig{
			Extra: parseKVPairs(setVars),
		},
	})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create gateway: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Gateway %q has been created!", gatewayName)

	return nil
}

func parseKVPairs(pairs []string) map[string]string {
	env := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		var v string
		if len(parts) > 1 {
			v = parts[1]
		}
		env[strings.TrimSpace(parts[0])] = v
	}
	return env
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Name,
		flags.ServerURL,
		flags.Credential,
	} {
		switch opt {
		case flags.Name:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("gateway name not configured")
			}
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Credential:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("credential not configured")
			}
		}
	}
	return nil
}
