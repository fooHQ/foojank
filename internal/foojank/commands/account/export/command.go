package export

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
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
	if c.Args().Len() < 1 {
		log.Error(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
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
