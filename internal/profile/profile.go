package profile

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
)

const (
	varOS                = "OS"
	varArch              = "ARCH"
	varTarget            = "TARGET"
	varFeatures          = "FEATURES"
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
	data profilesData
}

func ParseFile(file string) (*Profiles, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var data profilesData
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &Profiles{
		data: data,
	}, nil
}

func (p *Profiles) UnmarshalJSON(b []byte) error {
	var data profilesData
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	p.data = data
	return nil
}

func (p *Profiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.data)
}

func (p *Profiles) Get(name string) (*Profile, error) {
	var vd profileData
	if p.data.Defaults != nil {
		vd = *p.data.Defaults
	}

	v, ok := p.data.Profiles[name]
	if !ok {
		return nil, ErrProfileNotFound
	}

	return &Profile{
		data: merge(vd, v),
	}, nil
}

func (p *Profiles) Add(name string, profile *Profile, opt ...AddOption) error {
	var opts options
	for _, o := range opt {
		o(&opts)
	}

	if p.data.Profiles == nil {
		p.data.Profiles = make(map[string]profileData)
	}

	_, ok := p.data.Profiles[name]
	if ok && !opts.Overwrite {
		return ErrProfileExists
	}

	p.data.Profiles[name] = profile.data
	return nil
}

func (p *Profiles) Update(name string, profile *Profile) error {
	_, ok := p.data.Profiles[name]
	if !ok {
		return ErrProfileNotFound
	}

	p.data.Profiles[name] = profile.data
	return nil
}

func (p *Profiles) Delete(name string) error {
	_, ok := p.data.Profiles[name]
	if !ok {
		return ErrProfileNotFound
	}

	delete(p.data.Profiles, name)
	return nil
}

func (p *Profiles) List() []string {
	profs := make([]string, 0, len(p.data.Profiles))
	for name := range p.data.Profiles {
		profs = append(profs, name)
	}
	return profs
}

type Profile struct {
	data profileData
}

func NewProfile() *Profile {
	return &Profile{
		data: profileData{
			Environment: make(map[string]varData),
		},
	}
}

func Merge(profs ...*Profile) *Profile {
	data := make([]profileData, 0, len(profs))
	for _, prof := range profs {
		data = append(data, prof.data)
	}
	return &Profile{
		data: merge(data...),
	}
}

func (p *Profile) OS() string {
	if p.data.OS == nil {
		return ""
	}
	return *p.data.OS
}

func (p *Profile) Arch() string {
	if p.data.Arch == nil {
		return ""
	}
	return *p.data.Arch
}

func (p *Profile) Target() string {
	if p.data.Target == nil {
		return ""
	}
	return *p.data.Target
}

func (p *Profile) Features() []string {
	return p.data.Features
}

func (p *Profile) SetSourceDir(dir string) {
	p.data.SourceDir = new(dir)
}

func (p *Profile) Env() map[string]string {
	result := make(map[string]string, len(p.data.Environment))
	for k, v := range p.data.Environment {
		result[k] = v.Value
	}
	return result
}

func (p *Profile) SourceDir() string {
	if p.data.SourceDir == nil {
		return ""
	}
	return *p.data.SourceDir
}

func (p *Profile) Get(name string) *Var {
	v, ok := p.data.Environment[name]
	if !ok {
		return &Var{}
	}
	return &Var{
		data: v,
	}
}

func (p *Profile) SetOS(os string) {
	p.data.OS = new(os)
}

func (p *Profile) SetArch(arch string) {
	p.data.Arch = new(arch)
}

func (p *Profile) SetTarget(target string) {
	p.data.Target = new(target)
}

func (p *Profile) SetFeatures(features []string) {
	p.data.Features = slices.Clone(features)
}

func (p *Profile) Set(name string, v *Var) {
	p.data.Environment[name] = v.data
}

func (p *Profile) Delete(name string) {
	delete(p.data.Environment, name)
}

func (p *Profile) ToEnv() map[string]string {
	result := make(map[string]string, len(p.data.Environment))
	result[varOS] = ""
	if p.data.OS != nil {
		result[varOS] = *p.data.OS
	}
	result[varArch] = ""
	if p.data.Arch != nil {
		result[varArch] = *p.data.Arch
	}
	result[varTarget] = ""
	if p.data.Target != nil {
		result[varTarget] = *p.data.Target
	}
	result[varFeatures] = strings.Join(p.data.Features, ",")
	for k, v := range p.data.Environment {
		result[k] = v.Value
	}
	return result
}

func (p *Profile) UnmarshalJSON(b []byte) error {
	var data profileData
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if data.Environment == nil {
		data.Environment = make(map[string]varData)
	}
	p.data = data
	return nil
}

func (p *Profile) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.data)
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
	if v.data.Value != "" {
		return v.data.Value
	}
	return v.data.Default
}

func (v *Var) Description() string {
	return v.data.Description
}

func (v *Var) UnmarshalJSON(b []byte) error {
	var data varData
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	v.data = data
	return nil
}

func (v *Var) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.data)
}

type profilesData struct {
	Defaults *profileData           `json:"defaults,omitempty"`
	Profiles map[string]profileData `json:"profiles,omitempty"`
}

type profileData struct {
	OS          *string            `json:"os,omitempty"`
	Arch        *string            `json:"arch,omitempty"`
	Features    []string           `json:"features,omitempty"`
	Target      *string            `json:"target,omitempty"`
	SourceDir   *string            `json:"source_dir,omitempty"`
	Environment map[string]varData `json:"environment,omitempty"`
}

type varData struct {
	Value       string `json:"value"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

func merge(profs ...profileData) profileData {
	result := profileData{
		Environment: make(map[string]varData),
	}
	for _, prof := range profs {
		if prof.OS != nil {
			result.OS = prof.OS
		}
		if prof.Arch != nil {
			result.Arch = prof.Arch
		}
		if len(prof.Features) > 0 {
			result.Features = prof.Features
		}
		if prof.Target != nil {
			result.Target = prof.Target
		}
		if prof.SourceDir != nil {
			result.SourceDir = prof.SourceDir
		}
		for k, v := range prof.Environment {
			result.Environment[k] = varData{
				Value:       v.Value,
				Default:     v.Default,
				Description: v.Description,
			}
		}
	}
	return result
}

type options struct {
	Overwrite bool
}

type AddOption func(*options)

func WithOverwrite(overwrite bool) AddOption {
	return func(opts *options) {
		opts.Overwrite = overwrite
	}
}
