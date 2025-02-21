//go:build unix

package os

import (
	"net/url"
)

// ToURL converts path to `file://` URL. The function MUST NOT modify scheme. If the scheme is empty, it MUST remain empty.
func ToURL(path string) (*url.URL, error) {
	return url.Parse(path)
}
