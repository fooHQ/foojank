package runscript

import (
	"archive/zip"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/muesli/cancelreader"
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
		Usage:           "Execute Risor script locally",
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
		err := engineCompileAndRunPackage(ctx, pkgPath, pkgArgs)
		if err != nil && !errors.Is(err, context.Canceled) {
			err := fmt.Errorf("cannot execute a package: %w", err)
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			return err
		}

		return nil
	}
}

func engineCompileAndRunPackage(ctx context.Context, pkgPath string, pkgArgs []string) error {
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

	memHandler, err := engineos.NewMemURIHandler()
	if err != nil {
		err := fmt.Errorf("cannot create mem handler: %w", err)
		return err
	}
	defer memHandler.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	exitHandler := func(code int) {
		cancel()
	}

	stdin := engineos.NewPipe()
	stdout := engineos.NewPipe()
	r, err := cancelreader.NewReader(os.Stdin)
	if err != nil {
		err := fmt.Errorf("cannot create a standard input reader: %w", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(os.Stdout, stdout)
	}()

	o := engineos.New(
		engineos.WithArgs(pkgArgs),
		engineos.WithStdin(stdin),
		engineos.WithStdout(stdout),
		// Using URIFile with MemFS is intentional.
		// By default, runscript should not have access to the filesystem.
		// Work directory is also adjusted to begin at "/", which is the only
		// directory which exists in an empty MemFS.
		engineos.WithWorkDir("/"),
		engineos.WithURIHandler(engineos.URIFile, memHandler),
		engineos.WithExitHandler(exitHandler),
	)

	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		err := fmt.Errorf("cannot create zip reader: %w", err)
		return err
	}

	errCh := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = engine.Run(
			ctx,
			zr,
			engine.WithOS(o),
			engine.WithGlobals(config.Modules()),
			engine.WithGlobals(config.Builtins()),
		)
		_ = r.Cancel()
		errCh <- err
	}()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := append(scanner.Bytes(), '\n')
		_, _ = stdin.Write(line)
	}

	err = <-errCh
	_ = stdin.Close()
	_ = stdout.Close()
	wg.Wait()
	return err
}
