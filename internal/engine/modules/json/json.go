//go:build !module_json_stub

package json

import (
	modjson "github.com/risor-io/risor/modules/json"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modjson.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
