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
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
	"github.com/foohq/foojank/internal/foojank/path"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagFormat           = "format"
	FlagServer           = flags.Server
	FlagUserJWT          = flags.UserJWT
	FlagUserKey          = flags.UserKey
	FlagTLSCACertificate = flags.TLSCACertificate
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		ArgsUsage: "[repository:<file> ...]",
		Usage:     "List repositories or their contents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagFormat,
				Usage: "set output format",
				Value: "table",
			},
			&cli.StringSliceFlag{
				Name:  FlagServer,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  FlagUserJWT,
				Usage: "set user JWT token",
			},
			&cli.StringFlag{
				Name:  FlagUserKey,
				Usage: "set user secret key",
			},
			&cli.StringFlag{
				Name:  FlagTLSCACertificate,
				Usage: "set TLS CA certificate",
			},
		},
		Action:       action,
		Aliases:      []string{"ls"},
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	srv, err := server.New(conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey, *conf.Client.TLSCACertificate)
	if err != nil {
		log.Error(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	format := c.String(FlagFormat)

	if c.NArg() == 0 {
		err := listRepositories(ctx, srv, format)
		if err != nil {
			log.Error(ctx, "Cannot get a list of repositories: %v", err)
			return err
		}
		return nil
	}

	for _, pth := range c.Args().Slice() {
		result, err := parsePath(pth)
		if err != nil {
			log.Error(ctx, "Cannot parse path %q: %v", pth, err)
			return err
		}

		err = listRepository(ctx, srv, format, result.Repository, result.FilePath)
		if err != nil {
			log.Error(ctx, "Cannot list repository %q: %v", result.Repository, err)
			return err
		}
	}

	return nil
}

func listRepositories(ctx context.Context, srv *server.Client, format string) error {
	stores, err := srv.ListObjectStores(ctx)
	if err != nil {
		return err
	}

	table := formatter.NewTable([]string{
		"name",
		"size",
		"description",
	})
	for _, repo := range stores {
		name := repo.Name()
		size := formatBytes(repo.Size())
		description := repo.Description()
		table.AddRow([]string{
			name,
			size,
			description,
		})
	}

	return formatOutput(os.Stdout, format, table)
}

func listRepository(ctx context.Context, srv *server.Client, format, repository, pth string) error {
	store, err := srv.GetObjectStore(ctx, repository)
	if err != nil {
		return fmt.Errorf("cannot open repository: %w", err)
	}

	err = store.Wait(ctx)
	if err != nil {
		return fmt.Errorf("cannot synchronize repository: %w", err)
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

	err = formatOutput(os.Stdout, format, table)
	if err != nil {
		log.Error(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
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
			Repository: pth,
			FilePath:   "/",
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
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	if conf.Client == nil {
		return errors.New("client configuration is missing")
	}

	if len(conf.Client.Server) == 0 {
		return errors.New("server not configured")
	}

	if conf.Client.UserJWT == nil {
		return errors.New("user jwt not configured")
	}

	if conf.Client.UserKey == nil {
		return errors.New("user key not configured")
	}

	if conf.Client.TLSCACertificate == nil {
		return errors.New("tls ca certificate not configured")
	}

	return nil
}
