//go:build windows

package os_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/engine/os"
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
		u, err := os.ToURL(test.path)
		require.NoError(t, err)
		require.Equal(t, test.url, u.String(), "failed to convert path to URL (test %d/%d)", i+1, len(tests))
	}
}
