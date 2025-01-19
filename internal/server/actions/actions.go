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

func NewConfig(_ context.Context, c *cli.Command, validatorFn func(*config.Config) error) (*config.Config, error) {
	file := c.String(flags.Config)
	mustExist := c.IsSet(flags.Config)
	conf, err := config.ParseFile(file, mustExist)
	if err != nil {
		err = fmt.Errorf("cannot parse configuration file '%s': %w", file, err)
		return nil, err
	}

	if c.IsSet(flags.Host) {
		v := c.String(flags.Host)
		conf.Host = &v
	}

	if c.IsSet(flags.Port) {
		v := c.Int(flags.Port)
		conf.Port = &v
	}

	if c.IsSet(flags.OperatorJWT) {
		v := c.String(flags.OperatorJWT)
		conf.Operator = &config.Entity{
			JWT: v,
		}
	}

	if c.IsSet(flags.AccountJWT) {
		v := c.String(flags.AccountJWT)
		conf.Account = &config.Entity{
			JWT: v,
		}
	}

	if c.IsSet(flags.SystemAccountJWT) {
		v := c.String(flags.SystemAccountJWT)
		conf.SystemAccount = &config.Entity{
			JWT: v,
		}
	}

	if c.IsSet(flags.StoreDir) {
		v := c.String(flags.StoreDir)
		conf.StoreDir = &v
	}

	if c.IsSet(flags.LogLevel) {
		v := c.String(flags.LogLevel)
		conf.LogLevel = &v
	}

	if c.IsSet(flags.NoColor) {
		v := c.Bool(flags.NoColor)
		conf.NoColor = &v
	}

	if validatorFn != nil {
		err := validatorFn(conf)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}
