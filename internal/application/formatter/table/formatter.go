package table

import (
	"github.com/foojank/foojank/internal/application/formatter"
	"github.com/olekukonko/tablewriter"
	"io"
)

var _ formatter.Formatter = &Formatter{}

type Formatter struct{}

func New() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Write(o io.Writer, table *formatter.Table) error {
	w := tablewriter.NewWriter(o)
	w.SetHeader(table.Columns())
	w.AppendBulk(table.Rows())
	w.Render()
	return nil
}
