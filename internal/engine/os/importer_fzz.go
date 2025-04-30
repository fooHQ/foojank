package os

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/importer"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/parser"
)

var _ importer.Importer = &FzzImporter{}

// Extensions contains a list of supported script extensions.
var extensions = []string{
	".risor",
	".rsr",
}

type FzzImporter struct {
	reader    *zip.Reader
	opts      []compiler.Option
	codeCache map[string]*compiler.Code
	mux       sync.Mutex
}

func NewFzzImporter(reader io.ReaderAt, size int64, opts ...compiler.Option) (*FzzImporter, error) {
	r, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, err
	}

	return &FzzImporter{
		reader:    r,
		opts:      opts,
		codeCache: map[string]*compiler.Code{},
	}, nil
}

func (i *FzzImporter) Import(ctx context.Context, name string) (*object.Module, error) {
	i.mux.Lock()
	defer i.mux.Unlock()

	code, ok := i.codeCache[name]
	if ok {
		return object.NewModule(name, code), nil
	}

	var text string
	var found bool
	for _, ext := range extensions {
		text, found = i.readFile(name + ext)
		if found {
			break
		}
	}
	if !found {
		return nil, errors.New("import error: module \"" + name + "\" not found")
	}

	prog, err := parser.Parse(ctx, text)
	if err != nil {
		return nil, err
	}

	// TODO: add filename (compiler.WithFilename)
	code, err = compiler.Compile(prog, i.opts...)
	if err != nil {
		return nil, err
	}

	i.codeCache[name] = code

	return object.NewModule(name, code), nil
}

func (i *FzzImporter) readFile(name string) (string, bool) {
	file, err := i.reader.Open(name)
	if err != nil {
		return "", false
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return "", false
	}

	return string(b), true
}
