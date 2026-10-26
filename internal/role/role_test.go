package role_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/role"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roles.json")
	err := os.WriteFile(path, []byte(`[
		{"name":"operator","description":"operator role","privileges":["ROLE.CREATE","USER.LIST"],"created_at":"2026-01-01T00:00:00Z"},
		{"name":"viewer","description":"read only"}
	]`), 0o644)
	require.NoError(t, err)

	roles, err := role.ParseFile(path)
	require.NoError(t, err)
	require.Equal(t, []role.Role{
		{
			Name:        "operator",
			Description: "operator role",
			Privileges:  []string{"ROLE.CREATE", "USER.LIST"},
		},
		{
			Name:        "viewer",
			Description: "read only",
		},
	}, roles)
}

func TestParseFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roles.json")
	err := os.WriteFile(path, []byte(`[]`), 0o644)
	require.NoError(t, err)

	roles, err := role.ParseFile(path)
	require.NoError(t, err)
	require.Empty(t, roles)
}

func TestParseFileErrors(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		_, err := role.ParseFile(filepath.Join(dir, "missing.json"))
		require.Error(t, err)
		require.NotErrorIs(t, err, role.ErrParserError)
	})

	t.Run("invalid json", func(t *testing.T) {
		path := filepath.Join(dir, "bad.json")
		err := os.WriteFile(path, []byte(`{"roles":[]}`), 0o644)
		require.NoError(t, err)

		_, err = role.ParseFile(path)
		require.ErrorIs(t, err, role.ErrParserError)
	})

	t.Run("missing name", func(t *testing.T) {
		path := filepath.Join(dir, "noname.json")
		err := os.WriteFile(path, []byte(`[{"description":"no name"}]`), 0o644)
		require.NoError(t, err)

		_, err = role.ParseFile(path)
		require.EqualError(t, err, "role name is required")
	})

	t.Run("duplicate name", func(t *testing.T) {
		path := filepath.Join(dir, "dup.json")
		err := os.WriteFile(path, []byte(`[{"name":"operator"},{"name":"operator"}]`), 0o644)
		require.NoError(t, err)

		_, err = role.ParseFile(path)
		require.EqualError(t, err, `duplicate role "operator"`)
	})
}
