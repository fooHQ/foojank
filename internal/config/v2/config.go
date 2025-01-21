package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

var ErrParserError = errors.New("parser error")

type Config struct {
	*Common
	*Client
	*Server
}

func NewDefaultConfig() (*Config, error) {
	confCommon, err := NewDefaultCommon()
	if err != nil {
		return nil, err
	}

	confClient, err := NewDefaultClient()
	if err != nil {
		return nil, err
	}

	confServer, err := NewDefaultServer()
	if err != nil {
		return nil, err
	}

	return &Config{
		Common: confCommon,
		Client: confClient,
		Server: confServer,
	}, nil
}

func ParseFile(file string, v any) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = toml.Unmarshal(b, &conf)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &conf, nil
}

func ParseFlags(fn func(string) (any, bool)) (*Config, error) {
	confCommon, err := ParseCommonFlags(fn)
	if err != nil {
		return nil, err
	}

	confClient, err := ParseClientFlags(fn)
	if err != nil {
		return nil, err
	}
	// TODO: parseServerFlags

	return &Config{
		Common: confCommon,
		Client: confClient,
		Server: nil, // TODO!
	}, nil
}

func Merge(confs ...*Config) *Config {
	var result Config
	for _, conf := range confs {
		if conf == nil {
			continue
		}

		result.Common = MergeCommon(result.Common, conf.Common)
		result.Client = MergeClient(result.Client, conf.Client)
		// TODO: merge server!
	}
	return &result
}
