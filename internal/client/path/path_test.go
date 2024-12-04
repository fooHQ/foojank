package path

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected Path
		err      bool
	}{
		{"myrepo:/path/to/file", Path{"myrepo", "/path/to/file"}, false},
		{"anotherrepo:/another/path", Path{"anotherrepo", "/another/path"}, false},
		{"/path/to/file", Path{"", "/path/to/file"}, false},
		{"filename", Path{"", "filename"}, false},
		{"repo:", Path{}, true},
		{":/path/to/file", Path{}, true},
		{" : /path/to/file", Path{}, true},
		{"repo: ", Path{"repo", ""}, true},
		{" ", Path{}, true},
	}

	for _, test := range tests {
		result, err := Parse(test.input)
		if test.err {
			assert.NotNil(t, err)
			continue
		}
		assert.Equal(t, test.expected, result)
	}
}
