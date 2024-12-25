package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func CommandNotFound(_ context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("command '%s' does not exist", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}
