package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Common struct {
	DataDir  *string `toml:"data_dir,omitempty"`
	LogLevel *string `toml:"log_level,omitempty"`
	NoColor  *bool   `toml:"no_color,omitempty"`
}

func (c *Common) SetDataDir(dataDir string) {
	c.DataDir = &dataDir
}

func (c *Common) SetLogLevel(level string) {
	c.LogLevel = &level
}

func (c *Common) SetNoColor(noColor bool) {
	c.NoColor = &noColor
}

func NewDefaultCommon() (*Common, error) {
	dataDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var conf Common
	conf.SetDataDir(filepath.Join(dataDir, "foojank"))
	conf.SetLogLevel("info")
	conf.SetNoColor(false)
	return &conf, nil
}

func ParseCommonFlags(fn func(string) (any, bool)) (*Common, error) {
	var result Common
	configFields := map[string]func(string, any) error{
		"data_dir": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetDataDir(s)
			return nil
		},
		"log_level": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetLogLevel(s)
			return nil
		},
		"no_color": func(name string, v any) error {
			b, ok := v.(bool)
			if !ok {
				return fmt.Errorf("--%s must be a bool", name)
			}
			result.SetNoColor(b)
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

func MergeCommon(confs ...*Common) *Common {
	var result Common
	for _, conf := range confs {
		if conf == nil {
			continue
		}

		if conf.DataDir != nil {
			result.SetDataDir(*conf.DataDir)
		}

		if conf.LogLevel != nil {
			result.SetLogLevel(*conf.LogLevel)
		}

		if conf.NoColor != nil {
			result.SetNoColor(*conf.NoColor)
		}
	}
	return &result
}
