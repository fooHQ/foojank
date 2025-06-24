//go:build !module_os_stub

package os

import (
	modos "github.com/risor-io/risor/modules/os"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modos.Module()
}

func Builtins() map[string]object.Object {
	return modos.Builtins()
}
