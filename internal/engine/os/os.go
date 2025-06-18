package os

import (
	"errors"
	"os"
	"strings"

	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/urlpath"
)

var _ risoros.OS = &OS{}

type Option func(*OS)

func WithEnvVar(name, value string) Option {
	return func(o *OS) {
		o.environ[name] = value
	}
}

func WithArgs(args []string) Option {
	return func(o *OS) {
		o.args = args
	}
}

func WithStdin(file risoros.File) Option {
	return func(o *OS) {
		o.stdin = file
	}
}
func WithStdout(file risoros.File) Option {
	return func(o *OS) {
		o.stdout = file
	}
}

func WithWorkDir(dir string) Option {
	if dir == "" {
		dir = "/"
	}
	return func(o *OS) {
		o.wd = dir
	}
}

func WithFilesystems(fss map[string]risoros.FS) Option {
	return func(o *OS) {
		for scheme, fs := range fss {
			o.fs.registry[scheme] = fs
		}
	}
}

type ExitHandler func(int)

func WithExitHandler(handler ExitHandler) Option {
	return func(o *OS) {
		o.exitHandler = handler
	}
}

type OS struct {
	wd          string
	fs          *FS
	environ     map[string]string
	stdin       risoros.File
	stdout      risoros.File
	args        []string
	exitHandler ExitHandler
}

func (o *OS) Create(name string) (risoros.File, error) {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return nil, err
	}
	return o.fs.Create(pth)
}

func (o *OS) Mkdir(name string, perm os.FileMode) error {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return err
	}
	return o.fs.Mkdir(pth, perm)
}

func (o *OS) MkdirAll(path string, perm os.FileMode) error {
	pth, err := urlpath.Abs(path, o.wd)
	if err != nil {
		return err
	}
	return o.fs.MkdirAll(pth, perm)
}

func (o *OS) MkdirTemp(dir, pattern string) (string, error) {
	return "", errors.New("not implemented")
}

func (o *OS) Open(name string) (risoros.File, error) {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return nil, err
	}
	return o.fs.Open(pth)
}

func (o *OS) OpenFile(name string, flag int, perm os.FileMode) (risoros.File, error) {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return nil, err
	}
	return o.fs.OpenFile(pth, flag, perm)
}

func (o *OS) ReadFile(name string) ([]byte, error) {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return nil, err
	}
	return o.fs.ReadFile(pth)
}

func (o *OS) Remove(name string) error {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return err
	}
	return o.fs.Remove(pth)
}

func (o *OS) RemoveAll(path string) error {
	pth, err := urlpath.Abs(path, o.wd)
	if err != nil {
		return err
	}
	return o.fs.RemoveAll(pth)
}

func (o *OS) Rename(oldpath, newpath string) error {
	oldPth, err := urlpath.Abs(oldpath, o.wd)
	if err != nil {
		return err
	}
	newPth, err := urlpath.Abs(newpath, o.wd)
	if err != nil {
		return err
	}
	return o.fs.Rename(oldPth, newPth)
}

func (o *OS) Stat(name string) (os.FileInfo, error) {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return nil, err
	}
	return o.fs.Stat(pth)
}

func (o *OS) Symlink(oldname, newname string) error {
	oldPth, err := urlpath.Abs(oldname, o.wd)
	if err != nil {
		return err
	}
	newPth, err := urlpath.Abs(newname, o.wd)
	if err != nil {
		return err
	}
	return o.fs.Symlink(oldPth, newPth)
}

func (o *OS) TempDir() string {
	return os.TempDir()
}

func (o *OS) WriteFile(name string, content []byte, perm os.FileMode) error {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return err
	}
	return o.fs.WriteFile(pth, content, perm)
}

func (o *OS) ReadDir(name string) ([]risoros.DirEntry, error) {
	pth, err := urlpath.Abs(name, o.wd)
	if err != nil {
		return nil, err
	}
	return o.fs.ReadDir(pth)
}

func (o *OS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	pth, err := urlpath.Abs(root, o.wd)
	if err != nil {
		return err
	}
	return o.fs.WalkDir(pth, fn)
}

func (o *OS) PathSeparator() rune {
	return os.PathSeparator
}

func (o *OS) PathListSeparator() rune {
	return os.PathListSeparator
}

func (o *OS) Chdir(dir string) error {
	pth, err := urlpath.Abs(dir, o.wd)
	if err != nil {
		return err
	}

	info, err := o.fs.Stat(pth)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("chdir " + pth + ": file is not a directory")
	}

	scheme, err := urlpath.Scheme(pth)
	if err != nil {
		return err
	}

	if scheme == "file" {
		pth = strings.TrimPrefix(pth, "file://")
	}

	o.wd = pth
	return nil
}

func (o *OS) Getwd() (dir string, err error) {
	return o.wd, nil
}

func (o *OS) Stdout() risoros.File {
	return o.stdout
}

func (o *OS) Stdin() risoros.File {
	return o.stdin
}

func (o *OS) Args() []string {
	return o.args
}

func (o *OS) Environ() []string {
	var environ []string
	for k, v := range o.environ {
		environ = append(environ, k+"="+v)
	}
	return environ
}

func (o *OS) Getenv(key string) string {
	v, ok := o.environ[key]
	if !ok {
		return ""
	}
	return v
}

func (o *OS) Setenv(key, value string) error {
	o.environ[key] = value
	return nil
}

func (o *OS) Unsetenv(key string) error {
	delete(o.environ, key)
	return nil
}

func (o *OS) LookupEnv(key string) (string, bool) {
	v, ok := o.environ[key]
	return v, ok
}

func (o *OS) Exit(code int) {
	if o.exitHandler != nil {
		o.exitHandler(code)
	}
}

func (o *OS) Getpid() int {
	return os.Getpid()
}

func (o *OS) Getuid() int {
	return os.Getuid()
}

func (o *OS) Hostname() (string, error) {
	return os.Hostname()
}

func (o *OS) UserCacheDir() (string, error) {
	return os.UserCacheDir()
}

func (o *OS) UserConfigDir() (string, error) {
	return os.UserConfigDir()
}

func (o *OS) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (o *OS) CurrentUser() (risoros.User, error) {
	return risoros.Current()
}

func (o *OS) LookupUser(name string) (risoros.User, error) {
	return risoros.LookupUser(name)
}

func (o *OS) LookupUid(uid string) (risoros.User, error) {
	return risoros.LookupUid(uid)
}

func (o *OS) LookupGroup(name string) (risoros.Group, error) {
	return risoros.LookupGroup(name)
}

func (o *OS) LookupGid(gid string) (risoros.Group, error) {
	return risoros.LookupGid(gid)
}

func New(opts ...Option) *OS {
	o := &OS{
		wd:      initWD(),
		fs:      NewFS(),
		environ: initEnviron(),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func initWD() string {
	wd, _ := os.Getwd()
	if wd == "" {
		wd = "/"
	}
	return wd
}

func initEnviron() map[string]string {
	environ := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		environ[parts[0]] = parts[1]
	}
	return environ
}
