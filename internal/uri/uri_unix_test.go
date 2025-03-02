//go:build unix

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
			url:  "file:///home/user/test",
			path: "/home/user/test",
		},
		{
			url:  "/home/user/test",
			path: "/home/user/test",
		},
	}
	for i, test := range tests {
		u, err := url.Parse(test.url)
		require.NoError(t, err)
		p := uri.ToPath(u)
		require.Equal(t, test.path, p, "failed to convert URL to path (test %d/%d)", i+1, len(tests))
	}
}
