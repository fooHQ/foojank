package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

var ErrParserError = errors.New("parser error")

var (
	DefaultClientConfigPath = userConfigDir() + string(os.PathSeparator) + "client.conf"
)

type Config struct {
	*Common `toml:",inline"`
	Client  *Client `toml:"client"`
	Server  *Server `toml:"server"`
}

func (c *Config) String() string {
	b, _ := toml.Marshal(c)
	return string(b)
}

func (c *Config) Bytes() []byte {
	b, _ := toml.Marshal(c)
	return b
}

func NewDefault() (*Config, error) {
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

func ParseFile(file string) (*Config, error) {
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

	confServer, err := ParseServerFlags(fn)
	if err != nil {
		return nil, err
	}

	return &Config{
		Common: confCommon,
		Client: confClient,
		Server: confServer,
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
		result.Server = MergeServer(result.Server, conf.Server)
	}
	return &result
}

func userConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return dir + string(os.PathSeparator) + "foojank"
}
