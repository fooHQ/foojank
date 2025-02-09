package importers

import (
	"archive/zip"
	"context"
	"errors"
	"io"

	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/importer"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/parser"
)

var _ importer.Importer = &ZipImporter{}

// Extensions contains a list of supported script extensions.
var extensions = []string{
	".risor",
	".rsr",
}

type ZipImporter struct {
	reader *zip.Reader
	opts   []compiler.Option
}

func NewZipImporter(reader io.ReaderAt, size int64, opts ...compiler.Option) (*ZipImporter, error) {
	r, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, err
	}

	return &ZipImporter{
		reader: r,
		opts:   opts,
	}, nil
}

func (i *ZipImporter) Import(ctx context.Context, name string) (*object.Module, error) {
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

	code, err := compiler.Compile(prog, i.opts...)
	if err != nil {
		return nil, err
	}

	return object.NewModule(name, code), nil
}

func (i *ZipImporter) readFile(name string) (string, bool) {
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
