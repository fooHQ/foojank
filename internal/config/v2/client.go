package config

import (
	"fmt"
	"strings"
)

func WithServer(server []string) func(*Client) {
	return func(c *Client) {
		v := make([]string, len(server))
		copy(v, server)
		c.Server = v
	}
}

func WithUserJWT(jwt string) func(*Client) {
	return func(c *Client) {
		c.UserJWT = &jwt
	}
}

func WithUserKey(key string) func(*Client) {
	return func(c *Client) {
		c.UserKey = &key
	}
}

func WithAccountJWT(jwt string) func(*Client) {
	return func(c *Client) {
		c.AccountJWT = &jwt
	}
}

func WithAccountKey(key string) func(*Client) {
	return func(c *Client) {
		c.AccountKey = &key
	}
}

type Client struct {
	*Config
	Server     []string `toml:"server,omitempty"`
	UserJWT    *string  `toml:"user_jwt,omitempty"`
	UserKey    *string  `toml:"user_key,omitempty"`
	AccountJWT *string  `toml:"account_jwt,omitempty"`
	AccountKey *string  `toml:"account_key,omitempty"`
}

func NewDefaultClient() (*Client, error) {
	conf, err := NewDefaultConfig()
	if err != nil {
		return nil, err
	}

	return &Client{
		Config: conf,
		Server: []string{
			"ws://localhost",
		},
	}, nil
}

func ParseClientFile(file string) (*Client, error) {
	var conf *Client
	err := ParseFile(file, &conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func ParseClientFlags(fn func(string) (any, bool)) (*Client, error) {
	confBase, err := ParseFlags(fn)
	if err != nil {
		return nil, err
	}

	var result Client
	result.Config = confBase

	configFields := map[string]func(string, any) error{
		"server": func(name string, v any) error {
			s, ok := v.([]string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithServer(s)(&result)
			return nil
		},
		"user_jwt": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithUserJWT(s)(&result)
			return nil
		},
		"user_key": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithUserKey(s)(&result)
			return nil
		},
		"account_jwt": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithAccountJWT(s)(&result)
			return nil
		},
		"account_key": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			WithAccountKey(s)(&result)
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

func MergeClient(confs ...*Client) *Client {
	var result Client
	for _, conf := range confs {
		if conf == nil {
			continue
		}

		if conf.Config != nil {
			result.Config = Merge(result.Config, conf.Config)
		}

		if len(conf.Server) != 0 {
			WithServer(conf.Server)(&result)
		}

		if conf.UserJWT != nil {
			WithUserJWT(*conf.UserJWT)(&result)
		}

		if conf.UserKey != nil {
			WithUserKey(*conf.UserKey)(&result)
		}

		if conf.AccountJWT != nil {
			WithAccountJWT(*conf.AccountJWT)(&result)
		}

		if conf.AccountKey != nil {
			WithAccountKey(*conf.AccountKey)(&result)
		}
	}
	return &result
}
