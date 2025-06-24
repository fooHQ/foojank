//go:build module_net_stub

package net

import (
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return nil
}

func Builtins() map[string]object.Object {
	return nil
}
