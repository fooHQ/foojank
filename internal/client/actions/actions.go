package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/flags"
	"github.com/foohq/foojank/internal/config"
	// TODO: rename import!
	configv2 "github.com/foohq/foojank/internal/config/v2"
)

func NewClientConfig(_ context.Context, c *cli.Command) (*configv2.Client, error) {
	confDefault, err := configv2.NewDefaultClient()
	if err != nil {
		err = fmt.Errorf("cannot create a new configuration: %w", err)
		return nil, err
	}

	file := c.String(flags.Config)
	confFile, err := configv2.ParseClientFile(file)
	if err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("cannot parse configuration file '%s': %w", file, err)
		return nil, err
	}

	confFlags, err := configv2.ParseClientFlags(func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})

	result := configv2.MergeClient(confDefault, confFile, confFlags)
	return result, nil
}

// NewConfig TODO: can be deleted.
func NewConfig(_ context.Context, c *cli.Command, validatorFn func(*config.Config) error) (*config.Config, error) {
	file := c.String(flags.Config)
	mustExist := c.IsSet(flags.Config)
	conf, err := config.ParseFile(file, mustExist)
	if err != nil {
		err = fmt.Errorf("cannot parse configuration file '%s': %w", file, err)
		return nil, err
	}

	if c.IsSet(flags.Server) {
		conf.Servers = c.StringSlice(flags.Server)
	}

	if c.IsSet(flags.UserJWT) {
		if conf.User == nil {
			conf.User = &config.Entity{}
		}
		conf.User.JWT = c.String(flags.UserJWT)
	}

	if c.IsSet(flags.UserKey) {
		if conf.User == nil {
			conf.User = &config.Entity{}
		}
		conf.User.KeySeed = c.String(flags.UserKey)
	}

	if c.IsSet(flags.AccountJWT) {
		if conf.Account == nil {
			conf.Account = &config.Entity{}
		}
		conf.Account.JWT = c.String(flags.AccountJWT)
	}

	if c.IsSet(flags.AccountKey) {
		if conf.Account == nil {
			conf.Account = &config.Entity{}
		}
		conf.Account.SigningKeySeed = c.String(flags.AccountKey)
	}

	if c.IsSet(flags.LogLevel) {
		v := c.String(flags.LogLevel)
		conf.LogLevel = &v
	}

	if c.IsSet(flags.NoColor) {
		v := c.Bool(flags.NoColor)
		conf.NoColor = &v
	}

	if c.IsSet(flags.Codebase) {
		v := c.String(flags.Codebase)
		conf.Codebase = &v
	}

	if validatorFn != nil {
		err := validatorFn(conf)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func CommandNotFound(_ context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("command '%s' does not exist", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}
