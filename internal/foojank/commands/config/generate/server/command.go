package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/sstls"
)

const (
	FlagTLSOrganization = "tls-organization"
	FlagTLSDNSName      = "tls-dns-name"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Generate server configuration",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  FlagTLSDNSName,
				Usage: "set DNS names valid for TLS certificate",
			},
			&cli.StringFlag{
				Name:  FlagTLSOrganization,
				Usage: "set organization name on TLS certificate",
			},
		},
		Action: action,
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
	return createAction(logger)(ctx, c)
}

func createAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		dnsNames := c.StringSlice(FlagTLSDNSName)
		if len(dnsNames) == 0 {
			err := errors.New("cannot generate configuration: no DNS name provided for server's TLS certificate")
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		organization := c.String(FlagTLSOrganization)
		if organization == "" {
			organization = "ACME Co"
		}

		operatorName := fmt.Sprintf("OP%s", nuid.Next())
		operator, err := auth.NewOperator(operatorName)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountName := fmt.Sprintf("AC%s", nuid.Next())
		account, err := auth.NewAccount(accountName, []byte(operator.SigningKey), true)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		systemAccount, err := auth.NewAccount("SYS", []byte(operator.SigningKey), false)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		//
		// Generate server key and certificate
		//
		key, err := sstls.GenerateKey()
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		certTemplate, err := sstls.NewCertificateTemplate(organization, dnsNames)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		cert, err := sstls.EncodeCertificate(certTemplate, certTemplate, key.Public(), key)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		keyEncoded, err := sstls.EncodeKey(key)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confServer, err := config.NewDefaultServer()
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confServer.SetOperatorJWT(operator.JWT)
		confServer.SetOperatorKey(operator.Key)
		confServer.SetAccountJWT(account.JWT)
		// TODO: account's key is lost here!
		confServer.SetAccountKey(account.SigningKey)
		confServer.SetSystemAccountJWT(systemAccount.JWT)
		confServer.SetSystemAccountKey(systemAccount.Key)
		confServer.SetTLSCert(cert)
		confServer.SetTLSKey(keyEncoded)

		confCommon, err := config.NewDefaultCommon()
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confOutput := config.Config{
			Common: confCommon,
			Server: confServer,
		}
		_, _ = fmt.Fprintln(os.Stdout, confOutput.String())

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

	return nil
}
