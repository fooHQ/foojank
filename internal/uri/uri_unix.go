//go:build unix

package uri

import (
	"net/url"
	"path"
)

// ToURL converts path to a URL. The function MUST NOT modify scheme. If the scheme is empty, it MUST remain empty.
func ToURL(path string) (*url.URL, error) {
	return url.Parse(path)
}

func ToPath(u *url.URL) string {
	return u.Path
}

func ToFullPath(u *url.URL) string {
	if u.Scheme != "" && u.Scheme != "file" {
		return u.Scheme + "://" + u.Host + u.Path
	}

	return u.Path
}

func IsAbsoluteURL(u *url.URL) bool {
	return u.Scheme != "" || u.Host != ""
}

func NormalizeURL(wd, u *url.URL) *url.URL {
	if IsAbsoluteURL(u) {
		if u.Scheme == "" {
			u.Scheme = "file"
		}
		u.Path = path.Clean(u.Path)
	} else {
		u.Scheme = wd.Scheme
		if u.Scheme == "" {
			u.Scheme = "file"
		}
		u.Host = wd.Host
		if !path.IsAbs(u.Path) {
			u.Path = path.Join(wd.Path, u.Path)
		} else {
			u.Path = path.Clean(u.Path)
		}
	}

	return u
}
