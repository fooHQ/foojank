package formatter

import (
	"io"

	"github.com/deepnoodle-ai/wonton/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

var _ Formatter = (*ASCIIFormatter)(nil)

type ASCIIFormatter struct {
	opts options
}

func (f *ASCIIFormatter) Write(o io.Writer, table *Table) error {
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

	var rows [][]string
	for _, row := range table.Rows() {
		var newRow []string
		for _, cell := range row {
			newRow = append(newRow, f.formatCell(cell))
		}
		rows = append(rows, newRow)
	}

	err := w.Bulk(rows)
	if err != nil {
		return err
	}

	return w.Render()
}

func (f *ASCIIFormatter) formatCell(cell Cell) string {
	v := cell.String()
	if v == "" || f.opts.NoColor {
		return v
	}

	v = color.Colorize(cell.Color(), v)
	if cell.Bold() {
		v = color.ApplyBold(v)
	}

	return v
}
