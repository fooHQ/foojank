package engine

import (
	"context"
	"errors"

	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"
)

func Bootstrap(ctx context.Context) (*Code, error) {
	prog, err := parser.Parse(ctx, "import main")
	if err != nil {
		return nil, err
	}

	code, err := compiler.Compile(prog)
	if err != nil {
		return nil, err
	}

	return &Code{
		code: code,
	}, nil
}

type Code struct {
	code *compiler.Code
}

func (c *Code) Run(ctx context.Context, opts ...vm.Option) error {
	_, err := vm.Run(ctx, c.code, opts...)
	if err != nil {
		return &Error{err}
	}

	return nil
}

type Error struct {
	err error
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Error() string {
	var parserErr parser.ParserError
	if errors.As(e.err, &parserErr) {
		return parserErr.FriendlyErrorMessage()
	}
	return e.err.Error()
}
