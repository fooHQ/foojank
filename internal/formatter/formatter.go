package formatter

import (
	"io"
	"slices"
)

const (
	FormatJSON  = "json"
	FormatASCII = "ascii"
)

type Table struct {
	rows [][]Cell
}

func NewTable() *Table {
	return &Table{}
}

func (t *Table) AddRow(data []Cell) {
	t.rows = append(t.rows, slices.Clone(data))
}

func (t *Table) Rows() [][]Cell {
	return t.rows
}

type Formatter interface {
	Write(io.Writer, *Table) error
}

func NewFormatter(format string, opt ...Option) Formatter {
	var opts options
	for _, o := range opt {
		o(&opts)
	}
	switch format {
	case FormatJSON:
		return &JSONFormatter{
			opts: opts,
		}
	case FormatASCII:
		return &ASCIIFormatter{
			opts: opts,
		}
	default:
		return &ASCIIFormatter{
			opts: opts,
		}
	}
}

type options struct {
	NoColor bool
}

type Option func(*options)

func WithNoColor(noColor bool) Option {
	return func(opts *options) {
		opts.NoColor = noColor
	}
}
