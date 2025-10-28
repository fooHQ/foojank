package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/flags"
)

func NewConfig(_ context.Context, c *cli.Command) (*config.Config, error) {
	confFlags, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})
	if err != nil {
		err = fmt.Errorf("cannot parse command options: %w", err)
		return nil, err
	}

	configDir, isSet := confFlags.String(flags.ConfigDir)
	if !isSet {
		dir, err := FindConfigDir(".")
		if err != nil {
			return nil, errors.New("configuration directory not found in the current directory (or any of the parent directories)")
		}

		configDir = dir
	}

	isConfigDir, err := IsConfigDir(configDir)
	if err != nil {
		return nil, err
	}

	if !isConfigDir {
		err = fmt.Errorf("configuration directory not found in %q", configDir)
		return nil, err
	}

	confFile, err := ParseConfigJson(configDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("configuration directory not found in %q", configDir)
		} else {
			err = fmt.Errorf("cannot parse config file: %w", err)
		}
		return nil, err
	}

	conf := config.Merge(newDefaultConfig(), confFile, confFlags)
	return conf, nil
}

func newDefaultConfig() *config.Config {
	opts := map[string]any{
		flags.Format: "table",
	}

	if true { // TODO: check if output is tty!
		opts[flags.NoColor] = false
	}

	return config.NewWithOptions(opts)
}

func FindConfigDir(dir string) (string, error) {
	for i := 0; i < 128; i++ {
		isConfigDir, err := IsConfigDir(dir)
		if err != nil {
			return "", err
		}

		if !isConfigDir {
			dir = dir + "/../"
			continue
		}

		return dir, nil
	}

	return "", errors.New("configuration directory not found")
}

func InitConfigDir(dir string) error {
	err := os.MkdirAll(filepath.Join(dir, ".foojank"), 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	err = InitConfigJson(dir)
	if err != nil {
		return err
	}

	return nil
}

func IsConfigDir(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, ".foojank"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	if !info.IsDir() {
		return false, nil
	}

	_, err = ParseConfigJson(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func InitConfigJson(dir string) error {
	pth := filepath.Join(dir, ".foojank", "config.json")
	return os.WriteFile(pth, []byte("{}"), 0644)
}

func UpdateConfigJson(dir string, conf *config.Config) error {
	b, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	pth := filepath.Join(dir, ".foojank", "config.json")
	f, err := os.CreateTemp(filepath.Dir(pth), "config*.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(f.Name(), pth)
	if err != nil {
		return err
	}

	return nil
}

func ParseConfigJson(dir string) (*config.Config, error) {
	pth := filepath.Join(dir, ".foojank", "config.json")
	return config.ParseFile(pth)
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
