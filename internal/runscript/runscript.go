package runscript

import (
	"context"
	"fmt"
	"os"

	"github.com/risor-io/risor"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/engine"
	engineos "github.com/foohq/foojank/internal/engine/os"
	"github.com/foohq/foojank/internal/runscript/actions"
	"github.com/foohq/foojank/internal/runscript/config"
)

func New() *cli.Command {
	return &cli.Command{
		Name:            "runscript",
		ArgsUsage:       "<package>",
		Usage:           "Run Risor script locally",
		Version:         foojank.Version(),
		Action:          action,
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
		SkipFlagParsing: true,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	return runAction()(ctx, c)
}

func runAction() cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() < 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			_, _ = fmt.Fprintln(os.Stderr, err)
			return err
		}

		pkgPath := c.Args().First()
		pkgArgs := c.Args().Tail()
		err := executePackage(ctx, pkgPath, pkgArgs)
		if err != nil {
			err := fmt.Errorf("cannot execute a package: %w", err)
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			return err
		}

		return nil
	}
}

func executePackage(ctx context.Context, pkgPath string, pkgArgs []string) error {
	f, err := os.Open(pkgPath)
	if err != nil {
		err := fmt.Errorf("cannot open a package: %w", err)
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		err := fmt.Errorf("cannot stat a package: %w", err)
		return err
	}

	osCtx := engineos.NewContext(
		ctx,
		engineos.WithArgs(pkgArgs),
		engineos.WithStdin(os.Stdin),
		engineos.WithStdout(os.Stdout),
	)
	opts := []risor.Option{
		risor.WithoutDefaultGlobals(),
		risor.WithGlobals(config.Modules()),
		risor.WithGlobals(config.Builtins()),
	}
	c, err := engine.CompilePackage(osCtx, f, info.Size(), opts...)
	if err != nil {
		err := fmt.Errorf("cannot compile a package: %w", err)
		return err
	}

	return c.Run(osCtx)
}
