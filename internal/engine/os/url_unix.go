//go:build unix

package os

import (
	"net/url"
	"path"
)

func ToURL(path string) (*url.URL, error) {
	return url.Parse(path)
}

func NormalizeURL(wd, u *url.URL) *url.URL {
	if u.Scheme == "" {
		if wd.Scheme == "" {
			u.Scheme = "file"
		} else {
			u.Scheme = wd.Scheme
		}
	}

	if !path.IsAbs(u.Path) {
		u.Path = path.Join(wd.Path, u.Path)
	} else {
		u.Path = path.Clean(u.Path)
	}

	return u
}
