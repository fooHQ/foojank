package actions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/deepnoodle-ai/wonton/tty"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/configdir"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/profile"
)

func LoadConfig(w io.Writer, validateFn func(conf *config.Config) error) cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		newCtx, err := loadConfig(w, validateFn)(ctx, c)
		// urfave skips ShellComplete when Before fails. Flag names do not need a
		// working server or a project config, so keep completion running.
		if err != nil && IsShellCompletion() {
			return setConfigToContext(ctx, flagsOnlyConfig(c)), nil
		}
		return newCtx, err
	}
}

func loadConfig(w io.Writer, validateFn func(conf *config.Config) error) cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		if IsShellCompletion() {
			w = io.Discard
		}
		confFlags, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
			return c.Value(name), c.IsSet(name)
		})
		if err != nil {
			_, _ = fmt.Fprintf(w, "%s: cannot parse command options: %v\n", c.FullName(), err)
			return ctx, err
		}

		// A configuration directory is optional. Command-line options override the
		// file, and validateFn rejects the result when a required option is still
		// missing from both.
		confFile, configDir, err := readConfigFile(confFlags)
		if err != nil {
			_, _ = fmt.Fprintf(w, "%s: %v\n", c.FullName(), err)
			return ctx, err
		}

		conf := mergeConfig(configDir, confFile, confFlags)

		if !IsShellCompletion() {
			err = validateFn(conf)
			if err != nil {
				_, _ = fmt.Fprintf(w, "%s: invalid configuration: %v\n", c.FullName(), err)
				return ctx, err
			}
		}

		return setConfigToContext(ctx, conf), nil
	}
}

// readConfigFile loads .foojank/config.json. When no directory was requested
// and none can be found, the file is nil so configuration can come from
// command-line options alone. An explicit directory that does not exist, or a
// file that cannot be parsed, is an error.
func readConfigFile(confFlags *config.Config) (*config.Config, string, error) {
	configDir, explicit := confFlags.String(flags.ConfigDir)
	if !explicit {
		dir, err := configdir.Search(".")
		if err != nil {
			if errors.Is(err, configdir.ErrNotFound) {
				return nil, "", nil
			}
			return nil, "", configReadError(err)
		}
		configDir = dir
	}

	isConfigDir, err := configdir.IsConfigDir(configDir)
	if err != nil {
		return nil, "", configReadError(err)
	}
	if !isConfigDir {
		return nil, "", fmt.Errorf("configuration directory not found in %q", configDir)
	}

	confFile, err := configdir.ParseConfigJSON(configDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("configuration directory not found in %q", configDir)
		}
		return nil, "", fmt.Errorf("cannot parse config file: %w", err)
	}

	return confFile, configDir, nil
}

func configReadError(err error) error {
	if errors.Is(err, config.ErrParserError) {
		return fmt.Errorf("cannot parse config file: %w", err)
	}
	return err
}

func mergeConfig(configDir string, confFile, confFlags *config.Config) *config.Config {
	defs := map[string]string{
		flags.Format:  "ascii",
		flags.NoColor: "false",
	}
	if configDir != "" {
		defs[flags.ConfigDir] = configDir
	}

	confs := []*config.Config{
		config.NewWithOptions(defs),
		confFile,
		confFlags,
	}

	// If the output is not a TTY, disable color output.
	if !tty.IsTerminal(os.Stdout) {
		confs = append(confs, config.NewWithOptions(map[string]string{
			flags.NoColor: "true",
		}))
	}

	return config.Merge(confs...)
}

func flagsOnlyConfig(c *cli.Command) *config.Config {
	confFlags, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})
	if err != nil {
		confFlags = config.NewWithOptions(nil)
	}
	return config.Merge(config.NewWithOptions(map[string]string{
		flags.Format:  "ascii",
		flags.NoColor: "false",
	}), confFlags)
}

func LoadFlags(w io.Writer, validateFn func(conf *config.Config) error) cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		if IsShellCompletion() {
			w = io.Discard
		}
		conf, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
			return c.Value(name), c.IsSet(name)
		})
		if err != nil {
			err = fmt.Errorf("cannot parse command options: %w", err)
			_, _ = fmt.Fprintf(w, "%s: %v\n", c.FullName(), err)
			return ctx, err
		}

		if !IsShellCompletion() {
			err = validateFn(conf)
			if err != nil {
				_, _ = fmt.Fprintf(w, "%s: invalid configuration: %v\n", c.FullName(), err)
				return ctx, err
			}
		}

		return setConfigToContext(ctx, conf), nil
	}
}

func LoadProfiles(w io.Writer) cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		if IsShellCompletion() {
			w = io.Discard
		}
		conf := GetConfigFromContext(ctx)

		configDir, ok := conf.String(flags.ConfigDir)
		if !ok {
			if IsShellCompletion() {
				return ctx, nil
			}
			err := errors.New("cannot load profiles: configuration directory not set")
			_, _ = fmt.Fprintf(w, "%s: %v\n", c.FullName(), err)
			return ctx, err
		}

		profiles, err := configdir.ParseProfilesJSON(configDir)
		if err != nil {
			if IsShellCompletion() {
				return ctx, nil
			}
			err = fmt.Errorf("cannot parse profiles: %w", err)
			_, _ = fmt.Fprintf(w, "%s: %v\n", c.FullName(), err)
			return ctx, err
		}

		return setProfilesToContext(ctx, profiles), nil
	}
}

func SetupLogger(_ io.Writer) cli.BeforeFunc {
	return func(ctx context.Context, _ *cli.Command) (context.Context, error) {
		conf := GetConfigFromContext(ctx)

		noColor, ok := conf.Bool(flags.NoColor)
		if !ok {
			noColor = false
		}

		logger := log.NewLogger(log.LevelInfo, noColor)
		return setLoggerToContext(ctx, logger), nil
	}
}

func SetupServer(_ io.Writer) cli.BeforeFunc {
	return func(ctx context.Context, _ *cli.Command) (context.Context, error) {
		// urfave skips ShellComplete when Before fails. Flag names do not need
		// a live server connection.
		if IsShellCompletion() {
			return ctx, nil
		}

		conf := GetConfigFromContext(ctx)
		logger := GetLoggerFromContext(ctx)

		serverURL, _ := conf.String(flags.ServerURL)
		serverCert, _ := conf.String(flags.ServerCertificate)
		credsName, _ := conf.String(flags.Credential)

		credsFile, err := authdir.GetUserPath(credsName)
		if err != nil {
			if errors.Is(err, authdir.ErrUserNotFound) {
				err = fmt.Errorf("%q not found", credsName)
			}
			logger.ErrorContext(ctx, "Cannot read credential: %v", err)
			return ctx, err
		}

		srv, err := server.NewWithCredsFile([]string{serverURL}, credsFile, serverCert)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
			return ctx, err
		}

		return setServerToContext(ctx, srv), nil
	}
}

func UsageError(_ context.Context, c *cli.Command, err error, _ bool) error {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	return nil
}

func CommandNotFound(_ context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("%q is not a valid command", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}

type contextKey string

var (
	configKey   contextKey = "foojank:config"
	loggerKey   contextKey = "foojank:logger"
	profilesKey contextKey = "foojank:profiles"
	serverKey   contextKey = "foojank:server"
)

func GetConfigFromContext(ctx context.Context) *config.Config {
	// The function will panic if a context key is not found, that's intended to catch bugs early.
	conf := ctx.Value(configKey).(*config.Config)
	return conf
}

func GetLoggerFromContext(ctx context.Context) *log.Logger {
	logger := ctx.Value(loggerKey).(*log.Logger)
	return logger
}

func GetProfilesFromContext(ctx context.Context) *profile.Profiles {
	profs := ctx.Value(profilesKey).(*profile.Profiles)
	return profs
}

func GetServerFromContext(ctx context.Context) *server.Client {
	srv := ctx.Value(serverKey).(*server.Client)
	return srv
}

func setConfigToContext(ctx context.Context, conf *config.Config) context.Context {
	return context.WithValue(ctx, configKey, conf)
}

func setLoggerToContext(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func setProfilesToContext(ctx context.Context, profs *profile.Profiles) context.Context {
	return context.WithValue(ctx, profilesKey, profs)
}

func setServerToContext(ctx context.Context, srv *server.Client) context.Context {
	return context.WithValue(ctx, serverKey, srv)
}
