package os

import (
	"context"
	"os"
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

	o := &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		opts:     opts,
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
