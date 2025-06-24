//go:build !module_time_stub

package time

import (
	modtime "github.com/risor-io/risor/modules/time"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modtime.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
