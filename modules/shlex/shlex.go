package shlex

import (
	"context"

	"github.com/risor-io/risor/arg"
	"github.com/risor-io/risor/object"
	"github.com/u-root/u-root/pkg/shlex"
)

func Argv(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("shlex.argv", 1, args); err != nil {
		return err
	}

	data, err := object.AsString(args[0])
	if err != nil {
		return err
	}

	result := shlex.Argv(data)
	return object.NewStringList(result)
}

func Module() *object.Module {
	return object.NewBuiltinsModule("shlex", map[string]object.Object{
		"argv": object.NewBuiltin("argv", Argv),
	})
}

func Builtins() map[string]object.Object {
	return nil
}
