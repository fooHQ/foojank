package runner

import (
	"context"
	risoros "github.com/risor-io/risor/os"
)

type OS struct {
	*risoros.SimpleOS
	stdout risoros.File
}

func NewOS(ctx context.Context, stdout risoros.File) *OS {
	return &OS{
		SimpleOS: risoros.NewSimpleOS(ctx),
		stdout:   stdout,
	}
}

func (o *OS) Stdout() risoros.File {
	return o.stdout
}
