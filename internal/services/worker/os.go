package worker

import (
	"context"
	risoros "github.com/risor-io/risor/os"
)

type OS struct {
	*risoros.SimpleOS
	stdin  risoros.File
	stdout risoros.File
}

func NewOS(ctx context.Context, stdin, stdout risoros.File) *OS {
	return &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		stdin:    stdin,
		stdout:   stdout,
	}
}

func (o *OS) Stdout() risoros.File {
	return o.stdout
}

func (o *OS) Stdin() risoros.File {
	return o.stdin
}
