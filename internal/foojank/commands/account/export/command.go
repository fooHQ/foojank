package export

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:         "export",
		ArgsUsage:    "<account-name>",
		Usage:        "Export account's JWT",
		Action:       action,
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

	if c.Args().Len() < 1 {
		log.Error(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return err
	}

	name := c.Args().First()

	accountJWT, _, err := auth.ReadAccount(name)
	if err != nil {
		log.Error(ctx, "Cannot read account %q: %v", name, err)
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", accountJWT)

	return nil
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
