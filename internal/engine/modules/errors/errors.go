//go:build !module_errors_stub

package errors

import (
	moderrors "github.com/risor-io/risor/modules/errors"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return moderrors.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
