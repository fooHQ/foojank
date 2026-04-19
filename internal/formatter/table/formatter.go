package table

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	"github.com/foohq/foojank/internal/formatter"
)

var _ formatter.Formatter = &Formatter{}

type Formatter struct{}

func New() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Write(o io.Writer, table *formatter.Table) error {
	w := tablewriter.NewTable(
		o,
		tablewriter.WithRowConfig(tw.CellConfig{
			Padding: tw.CellPadding{
				Global: tw.PaddingDefault,
			},
			Filter: tw.CellFilter{
				Global: func(strings []string) []string {
					for i, s := range strings {
						if s == "" {
							strings[i] = "——"
						}
					}
					return strings
				},
			},
		}),
		tablewriter.WithRendition(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleDefault),
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		}))
	defer func() {
		_ = w.Close()
	}()
	w.Header(table.Columns())

	err := w.Bulk(table.Rows())
	if err != nil {
		return err
	}

	return w.Render()
}
