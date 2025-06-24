//go:build !module_filepath_stub

package filepath

import (
	modfilepath "github.com/risor-io/risor/modules/filepath"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modfilepath.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
