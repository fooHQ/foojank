package profile

import (
	"encoding/json"
	"errors"
	"os"
)

var ErrParserError = errors.New("parser error")

type Profiles struct {
	Profiles map[string]Profile `json:"profiles"`
	Schema   map[string]Schema  `json:"schema"`
}

type Profile struct {
	SourceDir   string            `json:"source_dir"`
	Environment map[string]string `json:"environment"`
}

type Schema struct {
	Default     string `json:"default"`
	Description string `json:"description"`
}

func ParseFile(file string) (*Profiles, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var data Profiles
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &data, nil
}
