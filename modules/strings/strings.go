package strings

import (
	modstrings "github.com/risor-io/risor/modules/strings"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modstrings.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
