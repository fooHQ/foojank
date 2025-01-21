package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/flags"
	// TODO: rename import!
	configv2 "github.com/foohq/foojank/internal/config/v2"
)

// NewClientConfig TODO: RENAME ME!
func NewClientConfig(_ context.Context, c *cli.Command) (*configv2.Config, error) {
	confDefault, err := configv2.NewDefault()
	if err != nil {
		err = fmt.Errorf("cannot create a new configuration: %w", err)
		return nil, err
	}

	file := c.String(flags.Config)
	confFile, err := configv2.ParseFile(file)
	if err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("cannot parse configuration file '%s': %w", file, err)
		return nil, err
	}

	confFlags, err := configv2.ParseFlags(func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})
	if err != nil {
		err = fmt.Errorf("cannot parse configuration flags: %w", err)
		return nil, err
	}

	result := configv2.Merge(confDefault, confFile, confFlags)
	return result, nil
}
