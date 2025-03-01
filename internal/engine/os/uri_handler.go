package os

import risoros "github.com/risor-io/risor/os"

type URIHandler interface {
	GetFS(host string) (risoros.FS, error)
}
