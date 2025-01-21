package config

import (
	"fmt"
	"strings"
)

type Client struct {
	Server     []string `toml:"server,omitempty"`
	UserJWT    *string  `toml:"user_jwt,omitempty"`
	UserKey    *string  `toml:"user_key,omitempty"`
	AccountJWT *string  `toml:"account_jwt,omitempty"`
	AccountKey *string  `toml:"account_key,omitempty"`
}

func (c *Client) SetServer(server []string) {
	v := make([]string, len(server))
	copy(v, server)
	c.Server = v
}

func (c *Client) SetUserJWT(jwt string) {
	c.UserJWT = &jwt
}

func (c *Client) SetUserKey(key string) {
	c.UserKey = &key
}

func (c *Client) SetAccountJWT(jwt string) {
	c.AccountJWT = &jwt
}

func (c *Client) SetAccountKey(key string) {
	c.AccountKey = &key
}

func NewDefaultClient() (*Client, error) {
	return &Client{
		Server: []string{
			"ws://localhost",
		},
	}, nil
}

func ParseClientFlags(fn func(string) (any, bool)) (*Client, error) {
	var result Client
	configFields := map[string]func(string, any) error{
		"server": func(name string, v any) error {
			s, ok := v.([]string)
			if !ok {
				return fmt.Errorf("--%s must be a string slice", name)
			}
			result.SetServer(s)
			return nil
		},
		"user_jwt": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetUserJWT(s)
			return nil
		},
		"user_key": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetUserKey(s)
			return nil
		},
		"account_jwt": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetAccountJWT(s)
			return nil
		},
		"account_key": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetAccountKey(s)
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

		if len(conf.Server) != 0 {
			result.SetServer(conf.Server)
		}

		if conf.UserJWT != nil {
			result.SetUserJWT(*conf.UserJWT)
		}

		if conf.UserKey != nil {
			result.SetUserKey(*conf.UserKey)
		}

		if conf.AccountJWT != nil {
			result.SetAccountJWT(*conf.AccountJWT)
		}

		if conf.AccountKey != nil {
			result.SetAccountKey(*conf.AccountKey)
		}
	}
	return &result
}
