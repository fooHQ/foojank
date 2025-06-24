//go:build !module_fmt_stub

package fmt

import (
	modfmt "github.com/risor-io/risor/modules/fmt"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modfmt.Module()
}

func Builtins() map[string]object.Object {
	return modfmt.Builtins()
}
