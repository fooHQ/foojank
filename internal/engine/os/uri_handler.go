package os

import (
	"net/url"

	risoros "github.com/risor-io/risor/os"
)

type URIHandler interface {
	GetFS(u *url.URL) (risoros.FS, error)
}
