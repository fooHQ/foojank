package master

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/sstls"
)

const (
	FlagForce = flags.Force
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "master",
		Usage: "Generate new master config",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    FlagForce,
				Usage:   "force overwrite a file if it already exists",
				Aliases: []string{"f"},
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
		force := c.Bool(FlagForce)

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

		// TODO: change name!
		certTemplate, err := sstls.NewCertificateTemplate("Test company")
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

		output, err := config.NewDefault()
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		output.Client.SetAccountJWT(account.JWT)
		output.Client.SetAccountKey(account.Key)
		output.Client.SetTLSCACert(cert)

		output.Server.SetOperatorJWT(operator.JWT)
		output.Server.SetOperatorKey(operator.Key)
		output.Server.SetAccountJWT(account.JWT)
		// TODO: account's key is lost here!
		output.Server.SetAccountKey(account.SigningKey)
		output.Server.SetSystemAccountJWT(systemAccount.JWT)
		output.Server.SetSystemAccountKey(systemAccount.Key)
		output.Server.SetTLSCert(cert)
		output.Server.SetTLSKey(keyEncoded)

		pth := config.DefaultMasterConfigPath
		dirPth := filepath.Dir(pth)
		if os.MkdirAll(dirPth, 0755) != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		opts := os.O_CREATE | os.O_WRONLY | os.O_EXCL
		if force {
			opts = opts &^ os.O_EXCL
		}

		f, err := os.OpenFile(pth, opts, 0600)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}
		defer f.Close()

		_, err = f.Write(output.Bytes())
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

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
