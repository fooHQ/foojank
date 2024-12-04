package config

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
)

var ErrParserError = errors.New("parser error")

type Entity struct {
	JWT string `yaml:"jwt"`
	Key string `yaml:"key"`
}

type Config struct {
	Servers  []string `yaml:"servers"`
	User     Entity   `yaml:"user"`
	LogLevel int64    `yaml:"log_level"`
	NoColor  bool     `yaml:"no_color"`
}

func (c *Config) Validate() error {
	return nil
}

func Parse(file string) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &conf, nil
}
