package base64

import (
	modbase64 "github.com/risor-io/risor/modules/base64"
	"github.com/risor-io/risor/object"
)

func Module() *object.Module {
	return modbase64.Module()
}

func Builtins() map[string]object.Object {
	return nil
}
