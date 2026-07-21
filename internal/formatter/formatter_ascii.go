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
			Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapBreak,
			},
			ColMaxWidths: tw.CellWidth{
				Global: 80,
			},
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

	var dataRows [][]Cell
	if header := table.Header(); header != nil {
		dataRows = append(dataRows, header)
	}
	dataRows = append(dataRows, sortedRows(table, f.opts)...)

	var rows [][]string
	for _, row := range dataRows {
		var newRow []string
		for _, cell := range row {
			newRow = append(newRow, f.formatCell(cell))
		}
		rows = append(rows, newRow)
	}

	if f.opts.Orientation == OrientationHorizontal {
		rows = transpose(rows)
	}

	err := w.Bulk(rows)
	if err != nil {
		return err
	}

	return w.Render()
}

// transpose rotates a row-major matrix so that column c becomes row c. The
// header (the first row) ends up as the leftmost column and each record's
// values are laid out next to it. Ragged rows are padded with empty cells.
func transpose(rows [][]string) [][]string {
	var cols int
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}

	out := make([][]string, cols)
	for c := 0; c < cols; c++ {
		col := make([]string, len(rows))
		for r, row := range rows {
			if c < len(row) {
				col[r] = row[c]
			}
		}
		out[c] = col
	}
	return out
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
