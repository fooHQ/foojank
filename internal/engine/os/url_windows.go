//go:build windows

package os

import (
	"net/url"
	"path/filepath"
	"strings"
)

func ToURL(path string) (*url.URL, error) {
	pth := strings.ReplaceAll(path, `\`, "/")
	volume := filepath.VolumeName(path)
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
