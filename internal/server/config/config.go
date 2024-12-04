package config

import (
	"github.com/goccy/go-yaml"
	"os"
)

type Entity struct {
	JWT string `yaml:"jwt"`
	Key string `yaml:"key"`
}

type Config struct {
	Host          string   `yaml:"host"`
	Operators     []Entity `yaml:"operators"`
	Accounts      []Entity `yaml:"accounts"`
	SystemAccount Entity   `yaml:"system_account"`
}

func Parse(file string) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// TODO: validate configuration
