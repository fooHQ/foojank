package engine

import (
	"context"
	"errors"
	"io"

	"github.com/risor-io/risor"
	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"

	"github.com/foohq/foojank/internal/engine/os"
)

func CompilePackage(ctx context.Context, reader io.ReaderAt, size int64, opts ...risor.Option) (*Code, error) {
	conf := risor.NewConfig(opts...)
	prog, err := parser.Parse(ctx, "import main")
	if err != nil {
		return nil, err
	}

	code, err := compiler.Compile(prog, conf.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	importer, err := os.NewFzzImporter(reader, size, conf.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	// Recreate risor config but include the custom importer this time.
	opts = append(opts, risor.WithImporter(importer))
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
	opts := risor.NewConfig(c.opts...)
	_, err := vm.Run(ctx, c.code, opts.VMOpts()...)
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
