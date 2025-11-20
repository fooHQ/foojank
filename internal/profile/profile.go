package profile

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
)

var (
	ErrParserError     = errors.New("parser error")
	ErrProfileNotFound = errors.New("profile not found")
	ErrProfileExists   = errors.New("profile already exists")
)

type Profiles struct {
	data profiles
}

func (p *Profiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.data)
}

func (p *Profiles) Get(name string) (Profile, error) {
	v, ok := p.data.Items[name]
	if !ok {
		return Profile{}, ErrProfileNotFound
	}
	return v, nil
}

func (p *Profiles) Add(name string, profile Profile) error {
	if p.data.Items == nil {
		p.data.Items = make(map[string]Profile)
	}

	_, ok := p.data.Items[name]
	if ok {
		return ErrProfileExists
	}

	p.data.Items[name] = profile
	return nil
}

func (p *Profiles) List() map[string]Profile {
	profs := make(map[string]Profile, len(p.data.Items))
	maps.Copy(profs, p.data.Items)
	return profs
}

type profiles struct {
	Items  map[string]Profile `json:"items"`
	Schema map[string]Schema  `json:"schema"`
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

	var data profiles
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &Profiles{
		data: data,
	}, nil
}
