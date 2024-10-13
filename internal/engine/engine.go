package engine

import (
	"context"
	"errors"
	"github.com/foojank/foojank/internal/engine/importers"
	"github.com/risor-io/risor"
	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"
	"io"
)

type Engine struct {
	opts     *risor.Config
	importer *importers.ZipImporter
	bootCode *compiler.Code
}

// TODO: do not use risor.Config as a parameter as it requires dependent modules to import risor as a dependency!
func Unpack(ctx context.Context, reader io.ReaderAt, size int64, opts *risor.Config) (*Engine, error) {
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

	return &Engine{
		opts:     opts,
		importer: imp,
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
