//go:build unix

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
			path: "/home/user/test",
			url:  "/home/user/test",
		},
		{
			path: "./home/user/test",
			url:  "./home/user/test",
		},
		{
			path: "../../home/user/test",
			url:  "../../home/user/test",
		},
		{
			path: "test",
			url:  "test",
		},
		{
			path: ".",
			url:  ".",
		},
		{
			path: "file:///home/user/test",
			url:  "file:///home/user/test",
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
			url:  "file:///home/user/test",
			path: "/home/user/test",
		},
		{
			url:  "/home/user/test",
			path: "/home/user/test",
		},
		{
			url:  "ftp://127.0.0.1/home/user/test",
			path: "ftp://127.0.0.1/home/user/test",
		},
	}
	for i, test := range tests {
		u, err := url.Parse(test.url)
		require.NoError(t, err)
		p := engineos.ToFullPath(u)
		require.Equal(t, test.path, p, "failed to convert URL to path (test %d/%d)", i+1, len(tests))
	}
}

func Test_ToPath(t *testing.T) {
	tests := []struct {
		url  string
		path string
	}{
		{
			url:  "file:///home/user/test",
			path: "/home/user/test",
		},
		{
			url:  "/home/user/test",
			path: "/home/user/test",
		},
		{
			url:  "ftp://127.0.0.1/home/user/test",
			path: "/home/user/test",
		},
	}
	for i, test := range tests {
		u, err := url.Parse(test.url)
		require.NoError(t, err)
		p := engineos.ToPath(u)
		require.Equal(t, test.path, p, "failed to convert URL to path (test %d/%d)", i+1, len(tests))
	}
}
