//go:build !module_net_stub

package net

import (
	modnet "github.com/risor-io/risor/modules/net"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modnet.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
