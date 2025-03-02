//go:build unix

package uri

import (
	"net/url"
)

func ToPath(u *url.URL) string {
	return u.Path
}
