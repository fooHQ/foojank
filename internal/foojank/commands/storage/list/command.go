package list

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
	"github.com/foohq/foojank/internal/foojank/path"
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
				Name:  flags.ServerCertificate,
				Usage: "set server TLS certificate",
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
		Before:       before,
		Action:       action,
		Aliases:      []string{"ls"},
		OnUsageError: actions.UsageError,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupLogger()(ctx, c)
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

	if c.NArg() == 0 {
		err := listStorages(ctx, srv, format)
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

		err = listStorage(ctx, srv, format, result.Storage, result.FilePath)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot list storage %q: %v", result.Storage, err)
			return err
		}
	}

	return nil
}

func listStorages(ctx context.Context, srv *server.Client, format string) error {
	stores, err := srv.ListObjectStores(ctx)
	if err != nil {
		return err
	}

	table := formatter.NewTable([]string{
		"name",
		"size",
		"description",
	})
	for _, storage := range stores {
		name := storage.Name()
		size := formatBytes(storage.Size())
		description := storage.Description()
		table.AddRow([]string{
			name,
			size,
			description,
		})
	}

	return formatOutput(os.Stdout, format, table)
}

func listStorage(ctx context.Context, srv *server.Client, format, storage, pth string) error {
	store, err := srv.GetObjectStore(ctx, storage)
	if err != nil {
		return fmt.Errorf("cannot open storage: %w", err)
	}

	err = store.Wait(ctx)
	if err != nil {
		return fmt.Errorf("cannot synchronize storage: %w", err)
	}

	info, err := store.Stat(pth)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		table := formatter.NewTable([]string{
			"type",
			"name",
			"size",
			"modified",
		})
		table.AddRow([]string{
			formatTypeIndicator(info.IsDir()),
			info.Mode().String(),
			info.Name(),
			formatBytes(uint64(info.Size())),
			formatTime(info.ModTime()),
		})
		return formatOutput(os.Stdout, format, table)
	}

	files, err := store.ReadDir(pth)
	if err != nil {
		return err
	}

	table := formatter.NewTable([]string{
		"type",
		"name",
		"size",
		"modified",
	})
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return err
		}

		table.AddRow([]string{
			formatTypeIndicator(info.IsDir()),
			info.Name(),
			formatBytes(uint64(info.Size())),
			formatTime(info.ModTime()),
		})
	}

	return formatOutput(os.Stdout, format, table)
}

func formatOutput(w io.Writer, format string, table *formatter.Table) error {
	var f formatter.Formatter
	switch format {
	case "json":
		f = jsonformatter.New()
	case "table":
		f = tableformatter.New()
	default:
		f = tableformatter.New()
	}

	err := f.Write(w, table)
	if err != nil {
		return fmt.Errorf("cannot write formatted output: %w", err)
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

func formatBytes(size uint64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota) // 1 << 10 = 1024
		MB
		GB
		TB
	)

	var unit string
	var value float64

	switch {
	case size >= TB:
		value = float64(size) / TB
		unit = "TB"
	case size >= GB:
		value = float64(size) / GB
		unit = "GB"
	case size >= MB:
		value = float64(size) / MB
		unit = "MB"
	case size >= KB:
		value = float64(size) / KB
		unit = "kB"
	default:
		value = float64(size)
		unit = "B"
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
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
