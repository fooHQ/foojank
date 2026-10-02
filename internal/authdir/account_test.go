package authdir_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/authdir"
)

func TestAccountDefaultRoot(t *testing.T) {
	setRoot(t, "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)

	token, seed := mustAccount(t, "main")
	require.NoError(t, authdir.WriteAccount("Main", token, seed))
	_, err = os.Stat(filepath.Join(configDir, "foojank", "accounts", "main", "account"))
	require.NoError(t, err)
}

func TestAccountRootOverride(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustAccount(t, "main")
	require.NoError(t, authdir.WriteAccount("Main", token, seed))
	_, err := os.Stat(filepath.Join(dir, "accounts", "main", "account"))
	require.NoError(t, err)

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, accounts)
}

func setRoot(t *testing.T, dir string) {
	t.Helper()
	prev := authdir.Root
	authdir.Root = dir
	t.Cleanup(func() {
		authdir.Root = prev
	})
}

func mustAccount(t *testing.T, name string) (string, []byte) {
	t.Helper()

	key, err := auth.NewAccountKey()
	require.NoError(t, err)

	claims, err := auth.NewAccountJWT(name, key)
	require.NoError(t, err)

	token, err := claims.Encode(key)
	require.NoError(t, err)

	seed, err := key.Seed()
	require.NoError(t, err)
	return token, seed
}
