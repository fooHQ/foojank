package os

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.OS = &OS{}

type Option func(*OS)

func WithEnvVar(name, value string) Option {
	return func(o *OS) {
		o.environ[name] = value
	}
}

func WithArgs(args ...string) Option {
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

type ExitHandler func(int)

func WithExitHandler(handler ExitHandler) Option {
	return func(o *OS) {
		o.exitHandler = handler
	}
}

type OS struct {
	*risoros.SimpleOS
	wd          string
	environ     map[string]string
	stdin       risoros.File
	stdout      risoros.File
	args        []string
	exitHandler ExitHandler
}

func (o *OS) Create(name string) (risoros.File, error) {
	pth := o.joinPath(name)
	return os.Create(pth)
}

func (o *OS) Mkdir(name string, perm os.FileMode) error {
	pth := o.joinPath(name)
	return os.Mkdir(pth, perm)
}

func (o *OS) MkdirAll(path string, perm os.FileMode) error {
	pth := o.joinPath(path)
	return os.MkdirAll(pth, perm)
}

func (o *OS) Open(name string) (risoros.File, error) {
	pth := o.joinPath(name)
	return os.Open(pth)
}

func (o *OS) OpenFile(name string, flag int, perm os.FileMode) (risoros.File, error) {
	pth := o.joinPath(name)
	return os.OpenFile(pth, flag, perm)
}

func (o *OS) ReadFile(name string) ([]byte, error) {
	pth := o.joinPath(name)
	return os.ReadFile(pth)
}

func (o *OS) Remove(name string) error {
	pth := o.joinPath(name)
	return os.Remove(pth)
}

func (o *OS) RemoveAll(path string) error {
	pth := o.joinPath(path)
	return os.RemoveAll(pth)
}

func (o *OS) Rename(oldpath, newpath string) error {
	oldPth := o.joinPath(oldpath)
	newPth := o.joinPath(newpath)
	return os.Rename(oldPth, newPth)
}

func (o *OS) Stat(name string) (os.FileInfo, error) {
	pth := o.joinPath(name)
	return os.Stat(pth)
}

func (o *OS) Symlink(oldname, newname string) error {
	oldPth := o.joinPath(oldname)
	newPth := o.joinPath(newname)
	return os.Symlink(oldPth, newPth)
}

func (o *OS) WriteFile(name string, content []byte, perm os.FileMode) error {
	pth := o.joinPath(name)
	return os.WriteFile(pth, content, perm)
}

func (o *OS) ReadDir(name string) ([]risoros.DirEntry, error) {
	pth := o.joinPath(name)
	results, err := os.ReadDir(pth)
	if err != nil {
		return nil, err
	}

	entries := make([]risoros.DirEntry, 0, len(results))
	for _, result := range results {
		entries = append(entries, &risoros.DirEntryWrapper{
			DirEntry: result,
		})
	}

	return entries, nil
}

func (o *OS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	pth := o.joinPath(root)
	return filepath.WalkDir(pth, fn)
}

func (o *OS) PathSeparator() rune {
	return os.PathSeparator
}

func (o *OS) PathListSeparator() rune {
	return os.PathListSeparator
}

func (o *OS) Chdir(dir string) error {
	pth := o.joinPath(dir)
	f, err := os.Open(pth)
	if err != nil {
		return err
	}
	defer f.Close()

	// Checks whether the file is a directory by trying to read the entries.
	_, err = f.Readdirnames(0)
	if err != nil {
		// Trying hard to return the same error string as stdlib's os.Chdir.
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			return errors.New("chdir " + pth + ": " + pathErr.Unwrap().Error())
		}
		return err
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

func (o *OS) joinPath(name string) string {
	if !filepath.IsAbs(name) {
		return filepath.Join(o.wd, name)
	}
	return name
}

func NewContext(ctx context.Context, options ...Option) context.Context {
	o := &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		wd:       initWD(),
		environ:  initEnviron(),
	}
	for _, option := range options {
		option(o)
	}

	return risoros.WithOS(ctx, o)
}

func initWD() string {
	wd, _ := os.Getwd()
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
