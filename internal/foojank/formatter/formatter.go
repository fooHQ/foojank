package formatter

import (
	"encoding/json"
	"io"
)

type Formatter interface {
	Write(io.Writer, *Table) error
}

type Table struct {
	header []string
	rows   [][]string
}

func NewTable(header []string) *Table {
	return &Table{
		header: header,
	}
}

func (t *Table) AddRow(data []string) {
	t.rows = append(t.rows, data)
}

func (t *Table) Columns() []string {
	return t.header
}

func (t *Table) Rows() [][]string {
	return t.rows
}

func (t *Table) MarshalJSON() ([]byte, error) {
	var o = struct {
		Header []string   `json:"header"`
		Rows   [][]string `json:"rows"`
	}{
		Header: t.header,
		Rows:   t.rows,
	}
	return json.Marshal(o)
}
