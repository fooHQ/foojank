package table

import (
	"io"

	"github.com/olekukonko/tablewriter"

	"github.com/foohq/foojank/internal/foojank/formatter"
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
