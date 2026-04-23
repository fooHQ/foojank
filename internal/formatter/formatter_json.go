package formatter

import (
	"encoding/json"
	"io"
)

var _ Formatter = (*JSONFormatter)(nil)

type JSONFormatter struct{}

func (f *JSONFormatter) Write(o io.Writer, table *Table) error {
	var data [][]any
	for _, row := range table.Rows() {
		var rowData []any
		for _, cell := range row {
			var v any
			switch c := cell.(type) {
			case StringCell:
				v = c.Value()
			case IntCell:
				v = c.Value()
			case UintCell:
				v = c.Value()
			case TimeCell:
				v = c.Value()
			}
			rowData = append(rowData, v)
		}
		data = append(data, rowData)
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = o.Write(b)
	if err != nil {
		return err
	}

	return nil
}
