package list

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/agent"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
	"github.com/foohq/foojank/internal/path"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		ArgsUsage: "[storage:<file> ...]",
		Usage:     "List storages or their contents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
			},
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set server account",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
		Aliases:         []string{"ls"},
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	userJWT, userSeed, err := auth.ReadUser(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", accountName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	client := agent.New(srv)

	if c.NArg() == 0 {
		err := listStorages(ctx, client, format, noColor)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get a list of storages: %v", err)
			return err
		}
		return nil
	}

	for _, pth := range c.Args().Slice() {
		result, err := parsePath(pth)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot parse path %q: %v", pth, err)
			return err
		}

		err = listStorage(ctx, client, format, noColor, result.Storage, result.FilePath)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot list storage %q: %v", result.Storage, err)
			return err
		}
	}

	return nil
}

func listStorages(ctx context.Context, client *agent.Client, format string, noColor bool) error {
	storages, err := client.ListStorage(ctx)
	if err != nil {
		return err
	}

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("SIZE").WithBold(),
		formatter.NewStringCell("DESCRIPTION").WithBold(),
	})
	for _, storage := range storages {
		status, err := storage.Status(ctx)
		if err != nil {
			return err
		}

		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(status.Name),
			formatter.NewSizeCell(status.Size),
			formatter.NewStringCell(status.Description),
		})
	}

	err = formatter.NewFormatter(format, formatter.WithNoColor(noColor)).Write(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}

func listStorage(ctx context.Context, client *agent.Client, format string, noColor bool, name, pth string) error {
	storageName, err := client.GetStorageName(ctx, name)
	if err != nil {
		return err
	}

	store, err := client.GetStorage(ctx, storageName)
	if err != nil {
		return err
	}

	err = store.Wait(ctx)
	if err != nil {
		return err
	}

	info, err := store.Stat(pth)
	if err != nil {
		return err
	}

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("TYPE").WithBold(),
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("SIZE").WithBold(),
		formatter.NewStringCell("MODIFIED").WithBold(),
	})

	if info.IsDir() {
		files, err := store.ReadDir(pth)
		if err != nil {
			return err
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				return err
			}

			table.AddRow([]formatter.Cell{
				formatter.NewStringCell(formatTypeIndicator(info.IsDir())),
				formatter.NewStringCell(info.Name()),
				formatter.NewSizeCell(uint64(info.Size())),
				formatter.NewTimeCell(info.ModTime()),
			})
		}
	} else {
		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(formatTypeIndicator(info.IsDir())),
			formatter.NewStringCell(info.Name()),
			formatter.NewSizeCell(uint64(info.Size())),
			formatter.NewTimeCell(info.ModTime()),
		})
	}

	err = formatter.NewFormatter(format, formatter.WithNoColor(noColor)).Write(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}

func parsePath(pth string) (path.Path, error) {
	if !strings.Contains(pth, ":") {
		return path.Path{
			Storage:  pth,
			FilePath: "/",
		}, nil
	}
	return path.Parse(pth)
}

func formatTypeIndicator(isDir bool) string {
	if isDir {
		return "DIR"
	}
	return "FILE"
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
