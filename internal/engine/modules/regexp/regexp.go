//go:build !module_regexp_stub

package regexp

import (
	modregexp "github.com/risor-io/risor/modules/regexp"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modregexp.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
