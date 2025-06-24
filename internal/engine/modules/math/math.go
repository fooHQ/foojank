//go:build !module_math_stub

package math

import (
	modmath "github.com/risor-io/risor/modules/math"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modmath.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
