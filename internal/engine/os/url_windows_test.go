//go:build windows

package os_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/engine/os"
)

func Test_ToURL(t *testing.T) {
	tests := []struct {
		path string
		url  *url.URL
	}{
		{
			path: `C:\Windows\System32`,
			url: &url.URL{
				Scheme: "file",
				Host:   "//C:",
				Path:   "/Windows/System32",
			},
		},
		{
			path: `C:/Windows/System32`,
			url: &url.URL{
				Scheme: "file",
				Host:   "//C:",
				Path:   "/Windows/System32",
			},
		},
		{
			path: `\\192.168.0.1\shared\data`,
			url: &url.URL{
				Scheme: "file",
				Host:   "//192.168.0.1/shared",
				Path:   "/data",
			},
		},
		{
			path: `\\192.168.0.1\shared`,
			url: &url.URL{
				Scheme: "file",
				Host:   "//192.168.0.1/shared",
				Path:   "/",
			},
		},
		{
			path: `//192.168.0.1/shared/data`,
			url: &url.URL{
				Scheme: "file",
				Host:   "//192.168.0.1/shared",
				Path:   "/data",
			},
		},
		{
			path: "file://C:/Windows/System32",
			url: &url.URL{
				Scheme: "file",
				Host:   "//C:",
			},
		},
	}
	for i, test := range tests {
		u, err := os.ToURL(test.path)
		require.NoError(t, err)
		require.Equal(t, test.url, u, "failed to convert path to URL (test %d/%d)", i+1, len(tests))
	}
}
