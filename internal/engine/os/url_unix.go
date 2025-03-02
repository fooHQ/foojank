//go:build unix

package os

import (
	"net/url"
)

// ToURL converts path to `file://` URL. The function MUST NOT modify scheme. If the scheme is empty, it MUST remain empty.
func ToURL(path string) (*url.URL, error) {
	return url.Parse(path)
}

func ToFullPath(u *url.URL) string {
	if u.Host != "" {
		if u.Scheme != "" && u.Scheme != "file" {
			return u.Scheme + "://" + u.Host + u.Path
		}
		return "//" + u.Host + u.Path
	}

	return u.Path
}

func ToPath(u *url.URL) string {
	return u.Path
}
