//go:build windows

package os_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	engineos "github.com/foohq/foojank/internal/engine/os"
)

func Test_ToURL(t *testing.T) {
	tests := []struct {
		path string
		url  string
	}{
		{
			path: `C:\Windows\System32`,
			url:  "/C:/Windows/System32",
		},
		{
			path: `C:/Windows/System32`,
			url:  "/C:/Windows/System32",
		},
		{
			path: `\\192.168.0.1\shared\data`,
			url:  "//192.168.0.1/shared/data",
		},
		{
			path: `\\192.168.0.1\shared`,
			url:  "//192.168.0.1/shared",
		},
		{
			path: `//192.168.0.1/shared/data`,
			url:  "//192.168.0.1/shared/data",
		},
		{
			path: `C:/Windows/System32`,
			url:  "/C:/Windows/System32",
		},
		{
			path: "file:///C:/Windows/System32",
			url:  "file:///C:/Windows/System32",
		},
		{
			path: "file://192.168.0.1/shared/data",
			url:  "file://192.168.0.1/shared/data",
		},
	}
	for i, test := range tests {
		u, err := engineos.ToURL(test.path)
		require.NoError(t, err)
		require.Equal(t, test.url, u.String(), "failed to convert path to URL (test %d/%d)", i+1, len(tests))
	}
}

func Test_ToFullPath(t *testing.T) {
	tests := []struct {
		url  string
		path string
	}{
		{
			url:  "/C:/Windows/System32",
			path: "C:/Windows/System32",
		},
		{
			url:  "//192.168.0.1/shared/data",
			path: "//192.168.0.1/shared/data",
		},
		{
			url:  "//192.168.0.1/shared",
			path: "//192.168.0.1/shared",
		},
		{
			url:  "file:///C:/Windows/System32",
			path: "C:/Windows/System32",
		},
		{
			url:  "file://192.168.0.1/shared/data",
			path: "//192.168.0.1/shared/data",
		},
		{
			url:  "ftp://192.168.0.2/shared/data",
			path: "ftp://192.168.0.2/shared/data",
		},
	}
	for i, test := range tests {
		u, err := url.Parse(test.url)
		require.NoError(t, err)
		p := engineos.ToFullPath(u)
		require.Equal(t, test.path, p, "failed to convert URL to path (test %d/%d)", i+1, len(tests))
	}
}

func Test_IsAbsoluteURL(t *testing.T) {
	// TODO
	_ = []struct {
		url    *url.URL
		result bool
	}{}
}

func Test_NormalizeURL(t *testing.T) {
	tests := []struct {
		workingDir *url.URL
		path       *url.URL
		result     *url.URL
	}{
		{
			workingDir: &url.URL{
				Scheme: "file",
				Path:   "/",
			},
			path: &url.URL{
				Path: "test",
			},
			result: &url.URL{
				Scheme: "file",
				Path:   "/test",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "file",
				Path:   "/",
			},
			path: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.1",
				Path:   "/test",
			},
			result: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.1",
				Path:   "/test",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.1",
				Path:   "/test",
			},
			path: &url.URL{
				Scheme: "file",
				Path:   "/var/log",
			},
			result: &url.URL{
				Scheme: "file",
				Path:   "/var/log",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.1",
				Path:   "/var/log",
			},
			path: &url.URL{
				Path: "../lib",
			},
			result: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.1",
				Path:   "/var/lib",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.1",
				Path:   "/var/log",
			},
			path: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.2",
				Path:   "/var/lib",
			},
			result: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.2",
				Path:   "/var/lib",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "file",
				Path:   "/C:/Users/user/Desktop",
			},
			path: &url.URL{
				Path: "../",
			},
			result: &url.URL{
				Scheme: "file",
				Path:   "/C:/Users/user",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "file",
				Host:   "192.168.0.1",
				Path:   "/shared",
			},
			path: &url.URL{
				Path: "../",
			},
			result: &url.URL{
				Scheme: "file",
				Host:   "192.168.0.1",
				Path:   "/",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "file",
				Host:   "192.168.0.1",
				Path:   "/shared",
			},
			path: &url.URL{
				Path: "/test",
			},
			result: &url.URL{
				Scheme: "file",
				Host:   "192.168.0.1",
				Path:   "/test",
			},
		},
		{
			workingDir: &url.URL{
				Scheme: "file",
				Host:   "192.168.0.1",
				Path:   "/shared/test",
			},
			path: &url.URL{
				Path: "/C:/Users/user/Desktop",
			},
			result: &url.URL{
				Scheme: "file",
				Path:   "/C:/Users/user/Desktop",
			},
		},
	}
	for i, test := range tests {
		result := engineos.NormalizeURL(test.workingDir, test.path)
		require.Equal(t, test.result, result, "failed to normalize URL (test %d/%d)", i+1, len(tests))
	}
}
