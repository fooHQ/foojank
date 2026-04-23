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

func NewFormatter(format string) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	case FormatASCII:
		return &ASCIIFormatter{}
	default:
		return &ASCIIFormatter{}
	}
}
