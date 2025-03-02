//go:build unix

package filefs

import (
	"net/url"
)

func ToPath(u *url.URL) string {
	return u.Path
}
