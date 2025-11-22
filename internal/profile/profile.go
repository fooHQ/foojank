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
	data map[string]Profile
}

func (p *Profiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.data)
}

func (p *Profiles) Get(name string) (Profile, error) {
	v, ok := p.data[name]
	if !ok {
		return Profile{}, ErrProfileNotFound
	}
	return v, nil
}

func (p *Profiles) Add(name string, profile Profile) error {
	if p.data == nil {
		p.data = make(map[string]Profile)
	}

	_, ok := p.data[name]
	if ok {
		return ErrProfileExists
	}

	p.data[name] = profile
	return nil
}

func (p *Profiles) Update(name string, profile Profile) error {
	_, ok := p.data[name]
	if !ok {
		return ErrProfileNotFound
	}

	p.data[name] = profile
	return nil
}

func (p *Profiles) Delete(name string) error {
	_, ok := p.data[name]
	if !ok {
		return ErrProfileNotFound
	}

	delete(p.data, name)
	return nil
}

func (p *Profiles) List() map[string]Profile {
	profs := make(map[string]Profile, len(p.data))
	maps.Copy(profs, p.data)
	return profs
}

type Profile struct {
	SourceDir   string         `json:"source_dir"`
	Environment map[string]Var `json:"environment"`
}

type Var struct {
	Value       string `json:"value"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

func ParseFile(file string) (*Profiles, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var data map[string]Profile
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &Profiles{
		data: data,
	}, nil
}
