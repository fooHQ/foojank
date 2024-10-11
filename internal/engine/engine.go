package engine

import (
	"context"
	"errors"
	"github.com/risor-io/risor"
	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"
)

type Engine struct {
	opts *risor.Config
}

func New() *Engine {
	return &Engine{
		opts: risor.NewConfig(),
	}
}

func (e *Engine) Build(ctx context.Context, input string) (*compiler.Code, error) {
	prog, err := parser.Parse(ctx, input)
	if err != nil {
		return nil, &Error{err}
	}

	code, err := compiler.Compile(prog, e.opts.CompilerOpts()...)
	if err != nil {
		return nil, &Error{err}
	}

	return code, nil
}

func (e *Engine) Eval(ctx context.Context, code *compiler.Code) (object.Object, error) {
	res, err := vm.Run(ctx, code, e.opts.VMOpts()...)
	if err != nil {
		return nil, &Error{err}
	}

	return res, nil
}

type Error struct {
	err error
}

func (e *Error) Error() string {
	var parserErr parser.ParserError
	if errors.As(e.err, &parserErr) {
		return parserErr.FriendlyErrorMessage()
	}
	return e.err.Error()
}
