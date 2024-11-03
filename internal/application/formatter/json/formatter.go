package json

import (
	"encoding/json"
	"github.com/foohq/foojank/internal/application/formatter"
	"io"
)

var _ formatter.Formatter = &Formatter{}

type Formatter struct{}

func New() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Write(o io.Writer, table *formatter.Table) error {
	b, err := json.Marshal(table)
	if err != nil {
		return err
	}

	_, err = o.Write(b)
	if err != nil {
		return err
	}

	return nil
}
