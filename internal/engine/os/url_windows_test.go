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

func Test_ToPath(t *testing.T) {
	tests := []struct {
		url  string
		path string
	}{
		{
			url:  "/C:/Windows/System32",
			path: "C:/Windows/System32",
		},
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
			url:  "//192.168.0.1/shared/data",
			path: "//192.168.0.1/shared/data",
		},
		{
			url:  "/C:/Windows/System32",
			path: "C:/Windows/System32",
		},
		{
			url:  "file:///C:/Windows/System32",
			path: "C:/Windows/System32",
		},
		{
			url:  "file://192.168.0.1/shared/data",
			path: "//192.168.0.1/shared/data",
		},
	}
	for i, test := range tests {
		u, err := url.Parse(test.url)
		require.NoError(t, err)
		p := engineos.ToPath(u)
		require.Equal(t, test.path, p, "failed to convert URL to path (test %d/%d)", i+1, len(tests))
	}
}
