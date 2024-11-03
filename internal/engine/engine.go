package engine

import (
	"context"
	"errors"
	"github.com/foohq/foojank/internal/engine/importers"
	"github.com/risor-io/risor"
	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"
	"io"
)

type Engine struct {
	opts     *risor.Config
	bootCode *compiler.Code
}

func Unpack(ctx context.Context, reader io.ReaderAt, size int64) (*Engine, error) {
	opts := risor.NewConfig()
	imp, err := importers.NewZipImporter(reader, size, opts.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	bootloader := "import main"
	prog, err := parser.Parse(ctx, bootloader)
	if err != nil {
		return nil, err
	}

	code, err := compiler.Compile(prog, opts.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	// Recreate risor config but include the custom importer this time.
	opts = risor.NewConfig(
		risor.WithImporter(imp),
	)

	return &Engine{
		opts:     opts,
		bootCode: code,
	}, nil
}

func (e *Engine) Run(ctx context.Context) (object.Object, error) {
	o, err := vm.Run(ctx, e.bootCode, e.opts.VMOpts()...)
	if err != nil {
		return nil, &Error{err}
	}

	return o, nil
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
