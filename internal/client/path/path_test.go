package path_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/client/path"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected path.Path
		err      bool
	}{
		{"myrepo:/path/to/file", path.Path{"myrepo", "/path/to/file"}, false},
		{"anotherrepo:/another/path", path.Path{"anotherrepo", "/another/path"}, false},
		{"/path/to/file", path.Path{"", "/path/to/file"}, false},
		{"filename", path.Path{"", "filename"}, false},
		{"repo:", path.Path{}, true},
		{":/path/to/file", path.Path{}, true},
		{" : /path/to/file", path.Path{}, true},
		{"repo: ", path.Path{"repo", ""}, true},
		{" ", path.Path{}, true},
	}

	for _, test := range tests {
		result, err := path.Parse(test.input)
		if test.err {
			require.NotNil(t, err)
			continue
		}
		require.Equal(t, test.expected, result)
	}
}
