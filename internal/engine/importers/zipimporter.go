package importers

import (
	"archive/zip"
	"context"
	"io"
	"io/fs"

	"github.com/risor-io/risor/compiler"
	"github.com/risor-io/risor/importer"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/parser"
)

var _ importer.Importer = &ZipImporter{}

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
	var file fs.File
	var err error
	for _, ext := range extensions {
		file, err = i.reader.Open(name + ext)
		if err == nil {
			break
		}
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	prog, err := parser.Parse(ctx, string(b))
	if err != nil {
		return nil, err
	}

	code, err := compiler.Compile(prog, i.opts...)
	if err != nil {
		return nil, err
	}

	return object.NewModule(name, code), nil
}
