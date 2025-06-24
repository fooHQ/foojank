//go:build !module_strconv_stub

package strconv

import (
	modstrconv "github.com/risor-io/risor/modules/strconv"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modstrconv.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
