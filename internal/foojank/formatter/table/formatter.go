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
	w.Header(table.Columns())

	err := w.Bulk(table.Rows())
	if err != nil {
		return err
	}

	return w.Render()
}
