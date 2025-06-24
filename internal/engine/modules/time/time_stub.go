//go:build module_time_stub

package time

import (
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return nil
}

func Builtins() map[string]object.Object {
	return nil
}
