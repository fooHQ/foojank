//go:build !module_http_stub

package http

import (
	modhttp "github.com/risor-io/risor/modules/http"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modhttp.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
