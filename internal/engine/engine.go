package engine

import (
	"context"
	"errors"

	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"
)

const entrypoint = `
from main import main

main()
`

func Run(ctx context.Context, opts ...vm.Option) error {
	prog, err := parser.Parse(ctx, entrypoint)
	if err != nil {
		return err
	}

	code, err := compiler.Compile(prog)
	if err != nil {
		return err
	}

	_, err = vm.Run(ctx, code, opts...)
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
