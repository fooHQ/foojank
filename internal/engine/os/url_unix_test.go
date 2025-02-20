//go:build unix

package os_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	engineos "github.com/foohq/foojank/internal/engine/os"
)

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
				Scheme: "file",
				Path:   "test",
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
				Path:   "test",
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
				Scheme: "smb",
				Host:   "192.168.0.1",
				Path:   "../lib",
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
				Path:   "../lib",
			},
			result: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.2",
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
				Path:   "../lib",
			},
			result: &url.URL{
				Scheme: "smb",
				Host:   "192.168.0.2",
				Path:   "/var/lib",
			},
		},
	}
	for _, test := range tests {
		result := engineos.NormalizeURL(test.workingDir, test.path)
		require.Equal(t, test.result, result)
	}
}
