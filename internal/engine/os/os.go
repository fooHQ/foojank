package os

import (
	"context"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.OS = &OS{}

type Options struct {
	stdin       risoros.File
	stdout      risoros.File
	args        []string
	exitHandler ExitHandler
}
type Option func(*Options)

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

func (o *OS) Exit(code int) {
	if o.opts.exitHandler != nil {
		o.opts.exitHandler(code)
	}
}

func NewContext(ctx context.Context, options ...Option) context.Context {
	var opts Options
	for _, option := range options {
		option(&opts)
	}

	o := &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		opts:     opts,
	}
	return risoros.WithOS(ctx, o)
}
