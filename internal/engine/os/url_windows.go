//go:build windows

package os

import (
	"net/url"
	"path/filepath"
	"strings"
)

func ToURL(path string) (*url.URL, error) {
	volume := filepath.VolumeName(path)
	// Is an absolute path starting with volume name (i.e. C:\)...
	if strings.Contains(volume, ":") {
		host := "//" + volume
		pth := strings.TrimPrefix(path, volume)
		pth = strings.ReplaceAll(pth, `\`, "/")
		return &url.URL{
			Scheme: "file",
			Host:   host,
			Path:   pth,
		}, nil
	}

	// Is UNC path...
	if strings.HasPrefix(volume, `\\`) {
		host := strings.ReplaceAll(volume, `\`, "/")
		pth := strings.TrimPrefix(path, volume)
		pth = strings.ReplaceAll(pth, `\`, "/")
		if pth == "" {
			pth = "/"
		}
		return &url.URL{
			Scheme: "file",
			Host:   host,
			Path:   pth,
		}, nil
	}

	path = strings.ReplaceAll(path, `\`, "/")
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	return u, nil
}
