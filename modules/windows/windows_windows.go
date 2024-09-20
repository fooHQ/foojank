package windows

import (
	"context"
	"fmt"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/op"
	"golang.org/x/sys/windows"
)

const LazyDLLType object.Type = "lazy_dll"

type LazyDLL struct {
	dll *windows.LazyDLL
}

func (d *LazyDLL) Type() object.Type {
	return LazyDLLType
}

func (d *LazyDLL) Inspect() string {
	return fmt.Sprintf("windows.lazy_dll(name: %s, system: %v)", d.dll.Name, d.dll.System)
}

func (d *LazyDLL) Interface() interface{} {
	return d.dll
}

func (d *LazyDLL) Equals(other object.Object) object.Object {
	if other.Type() != LazyDLLType {
		return object.NewBool(false)
	}
	return object.NewBool(d.dll == other.(*LazyDLL).dll)
}

func (d *LazyDLL) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "load":
		return object.NewBuiltin("windows.lazy_dll.load", func(ctx context.Context, args ...object.Object) object.Object {
			return d.dll.Load()
		}), true
	}
}

func (d *LazyDLL) SetAttr(name string, value object.Object) error {
	//TODO implement me
	panic("implement me")
}

func (d *LazyDLL) IsTruthy() bool {
	return true
}

func (d *LazyDLL) RunOperation(opType op.BinaryOpType, right object.Object) object.Object {
	return object.Errorf("eval error: unsupported operation for http.request: %v", opType)
}

func (d *LazyDLL) Cost() int {
	// I am not sure how to calculate the cost atm.
	return 8 + 8
}

func NewLazySystemDLL(ctx context.Context, args ...object.Object) object.Object {
	argsNum := len(args)
	if argsNum != 1 {
		return object.NewArgsError("windows.new_lazy_system_dll", 1, argsNum)
	}
	nameArg, err := object.AsString(args[0])
	if err != nil {
		return err
	}
	dll := windows.NewLazySystemDLL(nameArg)
	return &LazyDLL{
		dll: dll,
	}
}
