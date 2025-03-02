//go:build windows

package os

import (
	"net/url"
	"strings"
)

// ToURL converts path to `file://` URL. The function MUST NOT modify scheme. If the scheme is empty, it MUST remain empty.
func ToURL(path string) (*url.URL, error) {
	pth := strings.ReplaceAll(path, `\`, "/")
	if strings.Contains(pth, ":") {
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

func ToPath(u *url.URL) string {
	if strings.Contains(u.Path, ":") {
		return u.Path[1:]
	}

	if u.Host != "" {
		if u.Scheme != "" && u.Scheme != "file" {
			return u.Scheme + "://" + u.Host + u.Path
		}
		return "//" + u.Host + u.Path
	}

	return u.Path
}
