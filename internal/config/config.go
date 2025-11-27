package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
)

var ErrParserError = errors.New("parser error")

type config map[string]string

type Config struct {
	data config
}

func (c *Config) String(name string) (string, bool) {
	return c.get(name)
}

func (c *Config) Bool(name string) (bool, bool) {
	v, ok := c.get(name)
	if !ok {
		return false, false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, false
	}
	return b, true
}

func (c *Config) StringSlice(name string) ([]string, bool) {
	v, ok := c.get(name)
	if !ok {
		return nil, false
	}
	ss := strings.Split(v, ",")
	return ss, true
}

func (c *Config) get(name string) (string, bool) {
	v, ok := c.data[FlagToOption(name)]
	return v, ok
}

func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.data)
}

func NewWithOptions(opts map[string]string) *Config {
	data := make(config, len(opts))
	for k, v := range opts {
		data[FlagToOption(k)] = v
	}
	return &Config{
		data: data,
	}
}

func ParseFile(file string) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var data config
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &Config{
		data: data,
	}, nil
}

func ParseFlags(flags []string, fn func(string) (any, bool)) (*Config, error) {
	mdata := make(config, len(flags))
	for _, flag := range flags {
		v, ok := fn(flag)
		if !ok {
			continue
		}

		mdata[FlagToOption(flag)] = toString(v)
	}

	b, err := json.Marshal(mdata)
	if err != nil {
		return nil, err
	}

	var data config
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &Config{
		data: data,
	}, nil
}

func Merge(confs ...*Config) *Config {
	result := &Config{
		data: make(config),
	}
	for _, conf := range confs {
		if conf == nil {
			continue
		}

		for k, v := range conf.data {
			result.data[k] = v
		}
	}
	return result
}

func FlagToOption(flag string) string {
	return strings.ReplaceAll(flag, "-", "_")
}

func ParseKVPairs(pairs []string) map[string]string {
	env := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		var v string
		if len(parts) > 1 {
			v = parts[1]
		} else {
			v = ""
		}
		env[strings.TrimSpace(parts[0])] = v
	}
	return env
}

func toString(v any) string {
	switch vv := v.(type) {
	case string:
		return vv
	case []string:
		return strings.Join(vv, ",")
	case bool:
		return strconv.FormatBool(vv)
	default:
		return ""
	}
}
