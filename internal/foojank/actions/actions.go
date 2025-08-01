package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/flags"
)

func NewConfig(ctx context.Context, c *cli.Command) (*config.Config, error) {
	confDefault, err := config.NewDefault()
	if err != nil {
		err = fmt.Errorf("cannot create a new configuration: %w", err)
		return nil, err
	}

	confFile, file, err := parseConfigFile(ctx, c)
	if err != nil {
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

func parseConfigFile(_ context.Context, c *cli.Command) (*config.Config, string, error) {
	var file string
	if c.IsSet(flags.Config) {
		file = c.String(flags.Config)
	} else {
		file = config.DefaultClientConfigPath
	}

	conf, err := config.ParseFile(file)
	if err != nil {
		// The default config file does not exist, ignore the error.
		if os.IsNotExist(err) && !c.IsSet(flags.Config) {
			return nil, file, nil
		}

		return nil, file, err
	}

	return conf, file, nil
}

func UsageError(ctx context.Context, c *cli.Command, err error, _ bool) error {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	return nil
}

func CommandNotFound(_ context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("%q is not a valid command", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}
