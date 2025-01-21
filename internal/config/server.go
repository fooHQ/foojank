package config

import (
	"fmt"
	"strings"
)

type Server struct {
	Host             *string `toml:"host,omitempty"`
	Port             *int64  `toml:"port,omitempty"`
	OperatorJWT      *string `toml:"operator_jwt,omitempty"`
	OperatorKey      *string `toml:"operator_key,omitempty"`
	AccountJWT       *string `toml:"account_jwt,omitempty"`
	AccountKey       *string `toml:"account_key,omitempty"`
	SystemAccountJWT *string `toml:"system_account_jwt,omitempty"`
	SystemAccountKey *string `toml:"system_account_key,omitempty"`
}

func (s *Server) SetHost(host string) {
	s.Host = &host
}

func (s *Server) SetPort(port int64) {
	s.Port = &port
}

func (s *Server) SetOperatorJWT(jwt string) {
	s.OperatorJWT = &jwt
}

func (s *Server) SetOperatorKey(key string) {
	s.OperatorKey = &key
}

func (s *Server) SetAccountJWT(jwt string) {
	s.AccountJWT = &jwt
}

func (s *Server) SetAccountKey(key string) {
	s.AccountKey = &key
}

func (s *Server) SetSystemAccountJWT(jwt string) {
	s.SystemAccountJWT = &jwt
}

func (s *Server) SetSystemAccountKey(key string) {
	s.SystemAccountKey = &key
}

func NewDefaultServer() (*Server, error) {
	host := "0.0.0.0"
	port := int64(443)
	return &Server{
		Host: &host,
		Port: &port,
	}, nil
}

func ParseServerFlags(fn func(string) (any, bool)) (*Server, error) {
	var result Server
	configFields := map[string]func(string, any) error{
		"host": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetHost(s)
			return nil
		},
		"port": func(name string, v any) error {
			i, ok := v.(int64)
			if !ok {
				return fmt.Errorf("--%s must be an integer", name)
			}
			result.SetPort(i)
			return nil
		},
		"operator_jwt": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetOperatorJWT(s)
			return nil
		},
		"operator_key": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetOperatorKey(s)
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
		"system_account_jwt": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetSystemAccountJWT(s)
			return nil
		},
		"system_account_key": func(name string, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("--%s must be a string", name)
			}
			result.SetSystemAccountKey(s)
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

func MergeServer(confs ...*Server) *Server {
	var result Server
	for _, conf := range confs {
		if conf == nil {
			continue
		}

		if conf.Host != nil {
			result.SetHost(*conf.Host)
		}

		if conf.Port != nil {
			result.SetPort(*conf.Port)
		}

		if conf.OperatorJWT != nil {
			result.SetOperatorJWT(*conf.OperatorJWT)
		}

		if conf.OperatorKey != nil {
			result.SetOperatorKey(*conf.OperatorKey)
		}

		if conf.AccountJWT != nil {
			result.SetAccountJWT(*conf.AccountJWT)
		}

		if conf.AccountKey != nil {
			result.SetAccountKey(*conf.AccountKey)
		}

		if conf.SystemAccountJWT != nil {
			result.SetSystemAccountJWT(*conf.SystemAccountJWT)
		}

		if conf.SystemAccountKey != nil {
			result.SetSystemAccountKey(*conf.SystemAccountKey)
		}
	}
	return &result
}
