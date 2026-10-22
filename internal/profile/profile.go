package profile

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"strings"
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
			Commands:    make(map[string]commandData),
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

func (p *Profile) Env() map[string]string {
	result := make(map[string]string, len(p.data.Environment))
	for k, v := range p.data.Environment {
		result[k] = v.Value
	}
	return result
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

func (p *Profile) Set(name string, v *Var) {
	p.data.Environment[name] = v.data
}

func (p *Profile) Delete(name string) {
	delete(p.data.Environment, name)
}

func (p *Profile) Commands() map[string]*Command {
	out := make(map[string]*Command, len(p.data.Commands))
	for name, cmd := range p.data.Commands {
		cloned := cloneCommand(cmd)
		cloned.Name = name
		out[name] = &Command{data: cloned}
	}
	return out
}

func (p *Profile) GetCommand(name string) *Command {
	cmd, ok := p.data.Commands[name]
	if !ok {
		return &Command{}
	}
	cloned := cloneCommand(cmd)
	cloned.Name = name
	return &Command{data: cloned}
}

func (p *Profile) SetCommand(cmd *Command) {
	if cmd == nil {
		return
	}
	if p.data.Commands == nil {
		p.data.Commands = make(map[string]commandData)
	}
	p.data.Commands[cmd.data.Name] = cloneCommand(cmd.data)
}

func (p *Profile) DeleteCommand(name string) {
	delete(p.data.Commands, name)
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
	if data.Commands == nil {
		data.Commands = make(map[string]commandData)
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
	OS          *string                `json:"os,omitempty"`
	Arch        *string                `json:"arch,omitempty"`
	Environment map[string]varData     `json:"environment,omitempty"`
	Commands    map[string]commandData `json:"commands,omitempty"`
}

type Command struct {
	data commandData
}

func NewCommand(name string) *Command {
	return &Command{
		data: commandData{
			Name: name,
		},
	}
}

func (c *Command) Name() string {
	return c.data.Name
}

func (c *Command) Description() string {
	return c.data.Description
}

func (c *Command) Flags() map[string]*CommandFlag {
	out := make(map[string]*CommandFlag, len(c.data.Flags))
	for name, flag := range c.data.Flags {
		flag.Name = name
		out[name] = &CommandFlag{data: flag}
	}
	return out
}

func (c *Command) Environment() map[string]*Var {
	out := make(map[string]*Var, len(c.data.Environment))
	for name, v := range c.data.Environment {
		out[name] = &Var{data: v}
	}
	return out
}

func (c *Command) SetName(name string) {
	c.data.Name = name
}

func (c *Command) SetDescription(description string) {
	c.data.Description = description
}

func (c *Command) SetFlags(flags map[string]*CommandFlag) {
	if len(flags) == 0 {
		c.data.Flags = nil
		return
	}
	out := make(map[string]commandFlagData, len(flags))
	for name, flag := range flags {
		if flag == nil {
			continue
		}
		data := flag.data
		data.Name = name
		out[name] = data
	}
	if len(out) == 0 {
		c.data.Flags = nil
		return
	}
	c.data.Flags = out
}

func (c *Command) SetEnvironment(env map[string]*Var) {
	if len(env) == 0 {
		c.data.Environment = nil
		return
	}
	out := make(map[string]varData, len(env))
	for name, v := range env {
		if v == nil {
			continue
		}
		out[name] = v.data
	}
	if len(out) == 0 {
		c.data.Environment = nil
		return
	}
	c.data.Environment = out
}

func (c *Command) UnmarshalJSON(b []byte) error {
	var data commandData
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	c.data = data
	return nil
}

func (c *Command) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.data)
}

type CommandFlag struct {
	data commandFlagData
}

func NewCommandFlag(name string) *CommandFlag {
	return &CommandFlag{
		data: commandFlagData{
			Name: name,
		},
	}
}

func (f *CommandFlag) Name() string {
	return f.data.Name
}

func (f *CommandFlag) Description() string {
	return f.data.Description
}

func (f *CommandFlag) Default() string {
	return f.data.Default
}

func (f *CommandFlag) Required() bool {
	return f.data.Required
}

func (f *CommandFlag) SetName(name string) {
	f.data.Name = name
}

func (f *CommandFlag) SetDescription(description string) {
	f.data.Description = description
}

func (f *CommandFlag) SetDefault(value string) {
	f.data.Default = value
}

func (f *CommandFlag) SetRequired(required bool) {
	f.data.Required = required
}

func (f *CommandFlag) UnmarshalJSON(b []byte) error {
	var data commandFlagData
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	f.data = data
	return nil
}

func (f *CommandFlag) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.data)
}

type varData struct {
	Value       string `json:"value"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

type commandData struct {
	Name        string                     `json:"-"`
	Description string                     `json:"description,omitempty"`
	Flags       map[string]commandFlagData `json:"flags,omitempty"`
	Environment map[string]varData         `json:"environment,omitempty"`
}

type commandFlagData struct {
	Name        string `json:"-"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

func merge(profs ...profileData) profileData {
	result := profileData{
		Environment: make(map[string]varData),
		Commands:    make(map[string]commandData),
	}
	for _, prof := range profs {
		if prof.OS != nil {
			result.OS = prof.OS
		}
		if prof.Arch != nil {
			result.Arch = prof.Arch
		}
		for k, v := range prof.Environment {
			result.Environment[k] = varData{
				Value:       v.Value,
				Default:     v.Default,
				Description: v.Description,
			}
		}
		for name, cmd := range prof.Commands {
			cloned := cloneCommand(cmd)
			cloned.Name = name
			result.Commands[name] = cloned
		}
	}
	return result
}

func cloneCommand(cmd commandData) commandData {
	out := cmd
	if len(cmd.Flags) > 0 {
		out.Flags = make(map[string]commandFlagData, len(cmd.Flags))
		maps.Copy(out.Flags, cmd.Flags)
	}
	if len(cmd.Environment) > 0 {
		out.Environment = make(map[string]varData, len(cmd.Environment))
		maps.Copy(out.Environment, cmd.Environment)
	}
	return out
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
