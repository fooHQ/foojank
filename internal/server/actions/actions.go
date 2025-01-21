package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/server/flags"
)

func CommandNotFound(_ context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("command '%s' does not exist", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}

func NewConfig(_ context.Context, c *cli.Command) (*config.Config, error) {
	confDefault, err := config.NewDefault()
	if err != nil {
		err = fmt.Errorf("cannot create a new configuration: %w", err)
		return nil, err
	}

	file := c.String(flags.Config)
	confFile, err := config.ParseFile(file)
	if err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("cannot parse configuration file '%s': %w", file, err)
		return nil, err
	}

	confFlags, err := config.ParseFlags(func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})
	if err != nil {
		err = fmt.Errorf("cannot parse configuration flags: %w", err)
		return nil, err
	}

	result := config.Merge(confDefault, confFile, confFlags)
	return result, nil
}
