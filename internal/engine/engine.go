package engine

import (
	"context"
	"errors"
	"io"

	"github.com/risor-io/risor"
	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"

	"github.com/foohq/foojank/internal/engine/importers"
)

type Engine struct {
	opts []risor.Option
}

func New(opts ...risor.Option) *Engine {
	return &Engine{
		opts: opts,
	}
}

func (e *Engine) CompilePackage(ctx context.Context, reader io.ReaderAt, size int64) (*Code, error) {
	conf := risor.NewConfig(e.opts...)
	prog, err := parser.Parse(ctx, "import main")
	if err != nil {
		return nil, err
	}

	code, err := compiler.Compile(prog, conf.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	importer, err := importers.NewZipImporter(reader, size, conf.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	// Recreate risor config but include the custom importer this time.
	opts := append([]risor.Option{
		risor.WithImporter(importer),
	},
		e.opts...,
	)
	return &Code{
		code: code,
		opts: opts,
	}, nil
}

type Code struct {
	code *compiler.Code
	opts []risor.Option
}

func (c *Code) Run(ctx context.Context) error {
	conf := risor.NewConfig(c.opts...)
	_, err := vm.Run(ctx, c.code, conf.VMOpts()...)
	if err != nil {
		return &Error{err}
	}

	return nil
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
