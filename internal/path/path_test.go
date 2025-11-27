package path_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/path"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected path.Path
		err      bool
	}{
		{"mystorage:/path/to/file", path.Path{"mystorage", "/path/to/file"}, false},
		{"anotherstorage:/another/path", path.Path{"anotherstorage", "/another/path"}, false},
		{"/path/to/file", path.Path{"", "/path/to/file"}, false},
		{"filename", path.Path{"", "filename"}, false},
		{"storage:", path.Path{}, true},
		{":/path/to/file", path.Path{}, true},
		{" : /path/to/file", path.Path{}, true},
		{"storage: ", path.Path{"storage", ""}, true},
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
