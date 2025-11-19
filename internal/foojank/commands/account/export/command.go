package export

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:         "export",
		ArgsUsage:    "<account-name>",
		Usage:        "Export account's JWT",
		Before:       before,
		Action:       action,
		OnUsageError: actions.UsageError,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(io.Discard, validateConfiguration)(ctx, c)
	if err != nil {
		ctx, err = actions.LoadFlags(os.Stderr)(ctx, c)
		if err != nil {
			return ctx, err
		}
	}

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.GetLoggerFromContext(ctx)

	if c.Args().Len() < 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	accountJWT, _, err := auth.ReadAccount(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read account %q: %v", name, err)
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", accountJWT)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
