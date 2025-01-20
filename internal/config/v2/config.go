package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var ErrParserError = errors.New("parser error")

// TODO: can be removed!
type Entity struct {
	JWT            string `toml:"jwt"`
	KeySeed        string `toml:"key_seed,omitempty"`
	SigningKeySeed string `toml:"signing_key_seed,omitempty"`
}

func WithDataDir(dataDir string) func(*Config) {
	return func(c *Config) {
		c.DataDir = &dataDir
	}
}

func WithLogLevel(level string) func(*Config) {
	return func(c *Config) {
		c.LogLevel = &level
	}
}

func WithNoColor(noColor bool) func(*Config) {
	return func(c *Config) {
		c.NoColor = &noColor
	}
}

type Config struct {
	DataDir  *string `toml:"data_dir,omitempty"`
	LogLevel *string `toml:"log_level,omitempty"`
	NoColor  *bool   `toml:"no_color,omitempty"`
}

func NewDefaultConfig() (*Config, error) {
	dataDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var conf Config
	WithDataDir(filepath.Join(dataDir, "foojank"))(&conf)
	WithLogLevel("info")(&conf)
	WithNoColor(false)(&conf)
	return &conf, nil
}

func ParseFile(file string, v any) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(b, v)
	if err != nil {
		return errors.Join(ErrParserError, err)
	}

	return nil
}

func ParseFlags(fn func(string) (any, bool)) (*Config, error) {
	var result Config
	configFields := map[string]func(string, any) error{
		"data_dir": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithDataDir(s)(&result)
			return nil
		},
		"log_level": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithLogLevel(s)(&result)
			return nil
		},
		"no_color": func(name string, v any) error {
			b, ok := v.(bool)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithNoColor(b)(&result)
			return nil
		},
	}
	for fieldName, set := range configFields {
		flagName := strings.ReplaceAll(fieldName, "_", "-")
		v, ok := fn(flagName)
		if !ok {
			continue
		}

		err := set(flagName, v)
		if err != nil {
			return nil, err
		}
	}
	return &result, nil
}

func Merge(confs ...*Config) *Config {
	var result Config
	for _, conf := range confs {
		if conf == nil {
			continue
		}

		if conf.DataDir != nil {
			WithDataDir(*conf.DataDir)(&result)
		}

		if conf.LogLevel != nil {
			WithLogLevel(*conf.LogLevel)(&result)
		}

		if conf.NoColor != nil {
			WithNoColor(*conf.NoColor)(&result)
		}
	}
	return &result
}
