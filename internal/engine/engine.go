package engine

import (
	"context"
	"errors"
	"io/fs"

	"github.com/risor-io/risor"
	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/importer"
	risoros "github.com/risor-io/risor/os"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"
)

const entrypoint = `
from main import main

main()
`

var fileExtensions = []string{
	".risor",
	".rsr",
}

func Run(ctx context.Context, source fs.FS, opt ...Option) error {
	var opts Options
	for _, o := range opt {
		o(&opts)
	}

	conf := opts.toConfig()

	prog, err := parser.Parse(ctx, entrypoint)
	if err != nil {
		return err
	}

	code, err := compiler.Compile(prog)
	if err != nil {
		return err
	}

	imp := importer.NewFSImporter(importer.FSImporterOptions{
		GlobalNames: conf.GlobalNames(),
		SourceFS:    source,
		Extensions:  fileExtensions,
	})

	vmOpts := conf.VMOpts()
	vmOpts = append(vmOpts, vm.WithImporter(imp))
	_, err = vm.Run(ctx, code, vmOpts...)
	if err != nil {
		return &Error{err}
	}

	return nil
}

type Options struct {
	os      risoros.OS
	globals map[string]any
}

func (o *Options) Validate() error {
	if o.os == nil {
		return errors.New("engine: OS not specified")
	}

	return nil
}

func (o *Options) toConfig() *risor.Config {
	var opts = []risor.Option{
		risor.WithoutDefaultGlobals(),
	}

	if o.os != nil {
		opts = append(opts, risor.WithOS(o.os))
	}

	if o.globals != nil {
		opts = append(opts, risor.WithGlobals(o.globals))
	}

	return risor.NewConfig(opts...)
}

type Option func(*Options)

func WithOS(os risoros.OS) Option {
	return func(o *Options) {
		o.os = os
	}
}

func WithGlobals(globals map[string]any) Option {
	return func(o *Options) {
		if o.globals == nil {
			o.globals = make(map[string]any)
		}
		for k, v := range globals {
			o.globals[k] = v
		}
	}
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
