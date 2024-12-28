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

type Options struct {
	stdin       risoros.File
	stdout      risoros.File
	environ     map[string]string
	args        []string
	exitHandler ExitHandler
}
type Option func(*Options)

func WithEnvVar(name, value string) Option {
	return func(o *Options) {
		o.environ[name] = value
	}
}

func WithArgs(args ...string) Option {
	return func(o *Options) {
		o.args = args
	}
}

func WithStdin(file risoros.File) Option {
	return func(o *Options) {
		o.stdin = file
	}
}
func WithStdout(file risoros.File) Option {
	return func(o *Options) {
		o.stdout = file
	}
}

type ExitHandler func(int)

func WithExitHandler(handler ExitHandler) Option {
	return func(o *Options) {
		o.exitHandler = handler
	}
}

type OS struct {
	*risoros.SimpleOS
	opts Options
	wd   string
}

func (o *OS) Chdir(dir string) error {
	pth := dir
	if !filepath.IsAbs(dir) {
		pth = filepath.Join(o.wd, dir)
	}
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
			return errors.New("chdir " + dir + ": " + pathErr.Unwrap().Error())
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
	return o.opts.stdout
}

func (o *OS) Stdin() risoros.File {
	return o.opts.stdin
}

func (o *OS) Args() []string {
	return o.opts.args
}

func (o *OS) Environ() []string {
	var environ []string
	for k, v := range o.opts.environ {
		environ = append(environ, k+"="+v)
	}
	return environ
}

func (o *OS) Getenv(key string) string {
	v, ok := o.opts.environ[key]
	if !ok {
		return ""
	}
	return v
}

func (o *OS) Setenv(key, value string) error {
	o.opts.environ[key] = value
	return nil
}

func (o *OS) Unsetenv(key string) error {
	delete(o.opts.environ, key)
	return nil
}

func (o *OS) LookupEnv(key string) (string, bool) {
	v, ok := o.opts.environ[key]
	return v, ok
}

func (o *OS) Exit(code int) {
	if o.opts.exitHandler != nil {
		o.opts.exitHandler(code)
	}
}

func NewContext(ctx context.Context, options ...Option) context.Context {
	opts := Options{
		environ: initEnviron(),
	}
	for _, option := range options {
		option(&opts)
	}

	wd, _ := os.Getwd()
	o := &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		opts:     opts,
		wd:       wd,
	}
	return risoros.WithOS(ctx, o)
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
