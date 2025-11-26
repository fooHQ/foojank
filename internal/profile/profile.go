package profile

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

const (
	VarOS                = "OS"
	VarArch              = "ARCH"
	VarTarget            = "TARGET"
	VarFeatures          = "FEATURES"
	VarAgentID           = "FJ_AGENT_ID"
	VarServerURL         = "FJ_SERVER_URL"
	VarServerCertificate = "FJ_SERVER_CERTIFICATE"
	VarUserJWT           = "FJ_USER_JWT"
	VarUserKey           = "FJ_USER_KEY"
	VarStream            = "FJ_STREAM"
	VarConsumer          = "FJ_CONSUMER"
	VarInboxPrefix       = "FJ_INBOX_PREFIX"
	VarObjectStore       = "FJ_OBJECT_STORE"
)

var (
	ErrParserError     = errors.New("parser error")
	ErrProfileNotFound = errors.New("profile not found")
	ErrProfileExists   = errors.New("profile already exists")
)

type Profiles struct {
	data map[string]*Profile
}

func (p *Profiles) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &p.data)
}

func (p *Profiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.data)
}

func (p *Profiles) Get(name string) (*Profile, error) {
	v, ok := p.data[name]
	if !ok {
		return nil, ErrProfileNotFound
	}
	if v == nil {
		// If a profile with the name exists but is empty, return an empty profile.
		return New(), nil
	}

	return &Profile{
		data: v.data,
	}, nil
}

func (p *Profiles) Add(name string, profile *Profile) error {
	if p.data == nil {
		p.data = make(map[string]*Profile)
	}

	_, ok := p.data[name]
	if ok {
		return ErrProfileExists
	}

	p.data[name] = &Profile{
		data: profile.data,
	}
	return nil
}

func (p *Profiles) Update(name string, profile *Profile) error {
	_, ok := p.data[name]
	if !ok {
		return ErrProfileNotFound
	}

	p.data[name] = &Profile{
		data: profile.data,
	}
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

func (p *Profiles) List() []string {
	profs := make([]string, 0, len(p.data))
	for name := range p.data {
		profs = append(profs, name)
	}
	return profs
}

type Profile struct {
	data profileData
}

func New() *Profile {
	return &Profile{
		data: profileData{
			Environment: make(map[string]*Var),
		},
	}
}

func (p *Profile) SetSourceDir(dir string) {
	p.data.SourceDir = dir
}

func (p *Profile) SourceDir() string {
	return p.data.SourceDir
}

func (p *Profile) Get(name string) *Var {
	v, ok := p.data.Environment[name]
	if !ok {
		return &Var{}
	}
	return &Var{
		data: v.data,
	}
}

func (p *Profile) Set(name string, v *Var) {
	p.data.Environment[name] = &Var{
		data: v.data,
	}
}

func (p *Profile) Delete(name string) {
	delete(p.data.Environment, name)
}

func (p *Profile) List() map[string]string {
	result := make(map[string]string, len(p.data.Environment))
	for k, v := range p.data.Environment {
		result[k] = v.Value()
	}
	return result
}

func (p *Profile) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &p.data)
}

func (p *Profile) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.data)
}

type profileData struct {
	SourceDir   string          `json:"source_dir"`
	Environment map[string]*Var `json:"environment"`
}

type Var struct {
	data varData
}

func NewVar(value string) *Var {
	return &Var{
		data: varData{
			Value: value,
		},
	}
}

func ParseKVPairs(pairs []string) map[string]*Var {
	env := make(map[string]*Var, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		var v *Var
		if len(parts) > 1 {
			v = NewVar(parts[1])
		} else {
			v = NewVar("")
		}
		env[strings.TrimSpace(parts[0])] = v
	}
	return env
}

func (v *Var) SetValue(value string) {
	v.data.Value = value
}

func (v *Var) Value() string {
	if v.data.Value == "" {
		return v.data.Default
	}
	return v.data.Value
}

func (v *Var) Description() string {
	return v.data.Description
}

func (v *Var) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &v.data)
}

func (v *Var) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.data)
}

type varData struct {
	Value       string `json:"value"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

func ParseFile(file string) (*Profiles, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var data map[string]*Profile
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &Profiles{
		data: data,
	}, nil
}

func Merge(profs ...*Profile) *Profile {
	result := New()
	for _, prof := range profs {
		if prof == nil {
			continue
		}
		result.data.SourceDir = prof.data.SourceDir
		for k, v := range prof.data.Environment {
			result.data.Environment[k] = &Var{
				data: v.data,
			}
		}
	}
	return result
}
