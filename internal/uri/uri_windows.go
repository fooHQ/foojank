//go:build windows

package uri

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// ToURL converts path to a URL. The function MUST NOT modify scheme. If the scheme is empty, it MUST remain empty.
func ToURL(path string) (*url.URL, error) {
	pth := strings.ReplaceAll(path, `\`, "/")
	volume := filepath.VolumeName(pth)
	if strings.Contains(volume, ":") {
		// Is an absolute path starting with volume name (i.e. C:\)...
		return &url.URL{
			Path: "/" + pth,
		}, nil
	}

	u, err := url.Parse(pth)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// ToPath converts url to file path.
func ToPath(u *url.URL) string {
	if strings.Contains(u.Path, ":") {
		return u.Path[1:]
	}

	if (u.Scheme == "" || u.Scheme == "file") && u.Host != "" {
		return "//" + u.Host + u.Path
	}

	return u.Path
}

// ToFullPath converts url to full file path.
func ToFullPath(u *url.URL) string {
	if strings.Contains(u.Path, ":") {
		return u.Path[1:]
	}

	if u.Scheme != "" && u.Scheme != "file" {
		return u.Scheme + "://" + u.Host + path.Join("/", u.Path)
	}

	return "//" + u.Host + path.Join("/", u.Path)
}

func IsAbsoluteURL(u *url.URL) bool {
	return u.Scheme != "" || u.Host != "" || (u.Path != "" && filepath.VolumeName(u.Path[1:]) != "")
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
