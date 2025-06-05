//go:build unix

package urlpath_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/urlpath"
)

func TestAbs(t *testing.T) {
	tests := [][3]string{
		{"", "", "/"},
		{"test.txt", "", "/test.txt"},
		{"test.txt", "mem:///home/user", "mem:///home/user/test.txt"},
		{"../test.txt", "mem:///home/user", "mem:///home/test.txt"},
		{"/home/user", "mem:///test", "mem:///home/user"},
		{"/home/user", "mem:///test", "mem:///home/user"},
		{"file:///home/user", "mem:///test", "/home/user"},
		{"http://example.com/test.txt", "http://example.com:8888", "http://example.com/test.txt"},
		{"./", "mem:///test", "mem:///test"},
	}
	for i, test := range tests {
		abs, err := urlpath.Abs(test[0], test[1])
		require.NoError(t, err)
		require.Equal(t, test[2], abs, "test %d/%d failed", i+1, len(tests))
	}
}

func TestBase(t *testing.T) {
	tests := [][2]string{
		{"", "."},
		{"/", "/"},
		{"/home/user/test.txt", "test.txt"},
		{"./test.txt", "test.txt"},
		{"/home/user/../test.txt", "test.txt"},
		{"file:///home/user/test.txt", "test.txt"},
		{"mem:///home/user/test.txt", "test.txt"},
		{"http://localhost:8118/home/user/test.txt", "test.txt"},
		{"http://localhost:8118", "."},
	}
	for i, test := range tests {
		base, err := urlpath.Base(test[0])
		require.NoError(t, err)
		require.Equal(t, test[1], base, "test %d/%d failed", i+1, len(tests))
	}
}

func TestDir(t *testing.T) {
	tests := [][2]string{
		{"", "."},
		{"/", "/"},
		{"/home/user/test.txt", "/home/user"},
		{"./test.txt", "."},
		{"/home/user/../test.txt", "/home"},
		{"file:///home/user/test.txt", "/home/user"},
		{"mem:///home/user/test.txt", "mem:///home/user"},
		{"http://localhost:8118/home/user/test.txt", "http://localhost:8118/home/user"},
		// FIXME: behavior is inconsistent with Base.
		{"http://localhost:8118", "http://localhost:8118/"},
	}
	for i, test := range tests {
		base, err := urlpath.Dir(test[0])
		require.NoError(t, err)
		require.Equal(t, test[1], base, "test %d/%d failed", i+1, len(tests))
	}
}

func TestExt(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"/home/user/test.txt", ".txt"},
		{"./test.txt", ".txt"},
		{"/home/user/../test.txt", ".txt"},
		{"file:///home/user/test.txt", ".txt"},
		{"mem:///home/user/test.txt", ".txt"},
		{"http://localhost:8118/home/user/test.txt", ".txt"},
		{"http://localhost:8118", ""},
	}
	for i, test := range tests {
		base, err := urlpath.Ext(test[0])
		require.NoError(t, err)
		require.Equal(t, test[1], base, "test %d/%d failed", i+1, len(tests))
	}
}

func TestClean(t *testing.T) {
	tests := [][2]string{
		{"", "."},
		{"file:///home/user/../test.txt", "/home/test.txt"},
		{"mem:///home/user/test.txt", "mem:///home/user/test.txt"},
		{"mem:///home/user/./test.txt", "mem:///home/user/test.txt"},
		{"mem:///home/user/../test.txt", "mem:///home/test.txt"},
		{"http://localhost:8118/home/user/../test.txt", "http://localhost:8118/home/test.txt"},
		{"http://localhost:8118", "http://localhost:8118/"},
	}
	for i, test := range tests {
		clean, err := urlpath.Clean(test[0])
		require.NoError(t, err)
		require.Equal(t, test[1], clean, "test %d/%d failed", i+1, len(tests))
	}
}

func TestIsAbs(t *testing.T) {
	tests := map[string]bool{
		"":                           false,
		"/home/user/test.txt":        true,
		"./test.txt":                 false,
		"/home/user/../test.txt":     true,
		"file:///home/user/test.txt": true,
		"mem:///home/user/test.txt":  true,
		"http://localhost:8118/home/user/test.txt": true,
		"http://localhost:8118":                    true,
	}
	for name, is := range tests {
		base, err := urlpath.IsAbs(name)
		require.NoError(t, err)
		require.Equal(t, is, base, "test %q failed", name)
	}
}

func TestJoin(t *testing.T) {
	tests := [][3]string{
		{"mem:///home/user", "test.txt", "mem:///home/user/test.txt"},
		{"mem:///home/user", "mem:///test.txt", "mem:///home/user/test.txt"},
		{"/home/user", "mem:///test.txt", "/home/user/test.txt"},
	}
	for i, test := range tests {
		join, err := urlpath.Join(test[0], test[1])
		require.NoError(t, err)
		require.Equal(t, test[2], join, "test %d/%d failed", i+1, len(tests))
	}
}

func TestSplit(t *testing.T) {
	tests := [][3]string{
		{"mem:///home/user/test.txt", "mem:///home/user", "test.txt"},
		{"file:///home/user/test.txt", "/home/user", "test.txt"},
		{"http://localhost:8118/home/user/test.txt", "http://localhost:8118/home/user", "test.txt"},
	}
	for i, test := range tests {
		dir, file, err := urlpath.Split(test[0])
		require.NoError(t, err)
		require.Equal(t, test[1], dir, "test %d/%d failed", i+1, len(tests))
		require.Equal(t, test[2], file, "test %d/%d failed", i+1, len(tests))
	}
}

func TestMatch(t *testing.T) {
	tests := [][2]string{
		{"http://localhost:????/*.txt", "http://localhost:8118/test.txt"},
		{"*:///home/user/test.txt", "mem:///home/user/test.txt"},
		{"/home/user/*.txt", "/home/user/test.txt"},
	}
	for i, test := range tests {
		match, err := urlpath.Match(test[0], test[1])
		require.NoError(t, err)
		require.True(t, match, "test %d/%d failed", i+1, len(tests))
	}
}
