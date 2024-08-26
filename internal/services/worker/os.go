package worker

import (
	"context"
	risoros "github.com/risor-io/risor/os"
)

var _ risoros.OS = &OS{}

type OS struct {
	*risoros.SimpleOS
	ctx    context.Context
	cancel context.CancelFunc
	stdin  risoros.File
	stdout risoros.File
}

func NewOS(ctx context.Context, stdin, stdout risoros.File) *OS {
	ctx, cancel := context.WithCancel(ctx)
	return &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		ctx:      ctx,
		cancel:   cancel,
		stdin:    stdin,
		stdout:   stdout,
	}
}

func (o *OS) Context() context.Context {
	return risoros.WithOS(o.ctx, o)
}

func (o *OS) Stdout() risoros.File {
	return o.stdout
}

func (o *OS) Stdin() risoros.File {
	return o.stdin
}

func (o *OS) Exit(code int) {
	// TODO: preserve code!
	o.cancel()
}
