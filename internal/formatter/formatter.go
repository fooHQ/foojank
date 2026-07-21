package formatter

import (
	"cmp"
	"io"
	"slices"
)

const (
	FormatJSON  = "json"
	FormatASCII = "ascii"
)

type Table struct {
	header []Cell
	rows   [][]Cell
}

func NewTable() *Table {
	return &Table{}
}

func (t *Table) SetHeader(data []Cell) {
	t.header = slices.Clone(data)
}

func (t *Table) Header() []Cell {
	return t.header
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

// Orientation controls how a table is laid out when rendered.
type Orientation int

const (
	// OrientationVertical is the default layout: the header is the top row and
	// each record is a row beneath it.
	OrientationVertical Orientation = iota
	// OrientationHorizontal rotates the table to the left: the header is placed
	// in the leftmost column and each record's values are laid out next to it.
	OrientationHorizontal
)

type options struct {
	NoColor       bool
	SortByColumns []int
	Orientation   Orientation
}

type Option func(*options)

func WithNoColor(noColor bool) Option {
	return func(opts *options) {
		opts.NoColor = noColor
	}
}

func WithOrientation(orientation Orientation) Option {
	return func(opts *options) {
		opts.Orientation = orientation
	}
}

func WithSortByColumn(c ...int) Option {
	return func(opts *options) {
		opts.SortByColumns = c
	}
}

func sortedRows(table *Table, opts options) [][]Cell {
	rows := table.Rows()
	if len(opts.SortByColumns) == 0 {
		return rows
	}

	sorted := slices.Clone(rows)
	slices.SortStableFunc(sorted, func(a, b []Cell) int {
		for _, col := range opts.SortByColumns {
			if c := compareCells(cellAt(a, col), cellAt(b, col)); c != 0 {
				return c
			}
		}
		return 0
	})
	return sorted
}

func cellAt(row []Cell, c int) Cell {
	if c < 0 || c >= len(row) {
		return nil
	}
	return row[c]
}

// compareCells orders two cells, using their underlying values so numeric and
// time columns sort naturally rather than lexically. Cells of differing or
// unknown types fall back to comparing their string representation.
func compareCells(a, b Cell) int {
	switch av := a.(type) {
	case IntCell:
		if bv, ok := b.(IntCell); ok {
			return cmp.Compare(av.Value(), bv.Value())
		}
	case UintCell:
		if bv, ok := b.(UintCell); ok {
			return cmp.Compare(av.Value(), bv.Value())
		}
	case SizeCell:
		if bv, ok := b.(SizeCell); ok {
			return cmp.Compare(av.Value(), bv.Value())
		}
	case TimeCell:
		if bv, ok := b.(TimeCell); ok {
			return av.Value().Compare(bv.Value())
		}
	}

	var as, bs string
	if a != nil {
		as = a.String()
	}
	if b != nil {
		bs = b.String()
	}
	return cmp.Compare(as, bs)
}
