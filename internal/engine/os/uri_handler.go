package os

import (
	"net/url"

	risoros "github.com/risor-io/risor/os"
)

type URIHandler interface {
	GetFS(*url.URL) (risoros.FS, string, error)
	Close() error
}
