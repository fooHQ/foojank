//go:build windows

package uri

import (
	"net/url"
	"strings"
)

// ToPath converts url to file path.
func ToPath(u *url.URL) string {
	if strings.Contains(u.Path, ":") {
		return u.Path[1:]
	}

	if u.Host != "" {
		return "//" + u.Host + u.Path
	}

	return u.Path
}
