//go:build unix

package os

import (
	"net/url"
)

func ToURL(path string) (*url.URL, error) {
	return url.Parse(path)
}
