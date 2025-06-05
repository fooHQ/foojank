//go:build windows

package urlpath

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// fromString must not call path.Clean on URL.Path!
func fromString(pth string) (*url.URL, error) {
	pth = strings.ReplaceAll(pth, "\\", "/")
	volume := filepath.VolumeName(pth)
	if strings.Contains(volume, ":") {
		// Is an absolute path starting with volume name (i.e. C:\)...
		return &url.URL{
			Path: "/" + pth,
		}, nil
	}
	return url.Parse(pth)
}

func isAbsURL(u *url.URL) bool {
	return u.Scheme != "" || u.Host != "" || (u.Path != "" && filepath.VolumeName(u.Path[1:]) != "") || path.IsAbs(u.Path)
}

func toString(u *url.URL) string {
	if strings.Contains(u.Path, ":") {
		return path.Clean(u.Path[1:])
	}
	if !isFileURL(u) {
		return u.Scheme + "://" + u.Host + path.Join("/", u.Path)
	}
	if u.Host != "" {
		return "//" + u.Host + path.Join("/", u.Path)
	}
	return path.Clean(u.Path)
}

func normalize(u, wd *url.URL) string {
	if u.Scheme != "" || u.Host != "" || (u.Path != "" && filepath.VolumeName(u.Path[1:]) != "") {
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
	return toString(u)
}
