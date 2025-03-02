//go:build windows

package uri_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/uri"
)

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
	}
	for i, test := range tests {
		u, err := url.Parse(test.url)
		require.NoError(t, err)
		p := uri.ToPath(u)
		require.Equal(t, test.path, p, "failed to convert URL to path (test %d/%d)", i+1, len(tests))
	}
}
