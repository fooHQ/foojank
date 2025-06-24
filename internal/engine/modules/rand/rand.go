//go:build !module_rand_stub

package rand

import (
	modrand "github.com/risor-io/risor/modules/rand"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modrand.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
