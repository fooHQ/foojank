package os

import (
	"context"

	risoros "github.com/risor-io/risor/os"
)

var _ risoros.OS = &OS{}

type Options struct {
	stdin  risoros.File
	stdout risoros.File
	args   []string
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

type OS struct {
	*risoros.SimpleOS
	ctx    context.Context
	cancel context.CancelFunc
	opts   Options
}

func New(ctx context.Context, options ...Option) *OS {
	var opts Options
	for _, option := range options {
		option(&opts)
	}

	ctx, cancel := context.WithCancel(ctx)
	return &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		ctx:      ctx,
		cancel:   cancel,
		opts:     opts,
	}
}

func (o *OS) Context() context.Context {
	return risoros.WithOS(o.ctx, o)
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
	// TODO: preserve code!
	o.cancel()
}
