package create

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/auth"
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
			&cli.StringFlag{
				Name:  flags.URL,
				Usage: "set gateway URL",
			},
			&cli.StringFlag{
				Name:  flags.Certificate,
				Usage: "set path to gateway's certificate",
			},
			&cli.StringSliceFlag{
				Name:  flags.Extra,
				Usage: "set extra configuration (format: key=value)",
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

func action(ctx context.Context, c *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	gatewayName, _ := conf.String(flags.Name)
	gatewayDesc, _ := conf.String(flags.Description)
	gatewayURL, _ := conf.String(flags.URL)
	gatewayCert, _ := conf.String(flags.Certificate)
	gatewayExtra, _ := conf.StringSlice(flags.Extra)

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

	client := daemon.New(srv)

	var cert []byte
	if gatewayCert != "" {
		cert, err = readCertificateFile(gatewayCert)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot read certificate from %q: %v", gatewayCert, err)
			return err
		}
	}

	user, err := auth.NewUserKey()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user key: %v", err)
		return err
	}

	gatewayID, err := user.PublicKey()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create gateway ID: %v", err)
		return err
	}

	gatewayPerms := daemon.NewGatewayPermissions(gatewayID)
	gatewayClaims, err := auth.NewUserJWT(gatewayName, gatewayPerms, user)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return err
	}

	// Get the client's user JWT.
	userClaims, err := auth.GetUserJWT(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get user JWT: %v", err)
		return err
	}

	if userClaims.IssuerAccount != "" {
		gatewayClaims.IssuerAccount = userClaims.IssuerAccount
	}

	// Get the client's account key.
	account, err := auth.GetAccountKey(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get account key: %v", err)
		return err
	}

	gatewayJWT, err := gatewayClaims.Encode(account)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return err
	}

	gatewaySeed, err := user.Seed()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode user seed: %v", err)
		return err
	}

	gateway := daemon.GatewayDirectoryEntry{
		ID:          gatewayID,
		Name:        gatewayName,
		Description: gatewayDesc,
		Config: daemon.GatewayConfig{
			URL:         gatewayURL,
			Certificate: cert,
			UserJWT:     gatewayJWT,
			UserKey:     string(gatewaySeed),
			Extra:       parseKVPairs(gatewayExtra),
		},
	}
	err = client.RegisterGateway(ctx, gateway)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot register gateway: %v", err)
		return err
	}

	err = client.CreateGateway(ctx, gateway)
	if err != nil {
		if errors.Is(err, daemon.ErrKeyExists) {
			err = fmt.Errorf("%q already exists", gatewayName)
		}
		logger.ErrorContext(ctx, "Cannot create gateway: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Gateway %q has been built!", gatewayName)

	return nil
}

func readCertificateFile(pth string) ([]byte, error) {
	b, err := os.ReadFile(pth)
	if err != nil {
		return nil, err
	}

	for {
		var block *pem.Block
		block, b = pem.Decode(b)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		return cert.Raw, nil
	}

	return nil, errors.New("no certificate PEM block found")
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
		flags.URL,
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
		case flags.Name:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("gateway name not configured")
			}
		case flags.URL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("gateway URL name not configured")
			}
			u, err := url.ParseRequestURI(v)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return errors.New("gateway URL format is invalid")
			}
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
