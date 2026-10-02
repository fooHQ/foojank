package authdir_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nats-io/jwt/v2"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/authdir"
)

func TestUserDefaultRoot(t *testing.T) {
	setRoot(t, "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)

	token, seed := mustUser(t, "main")
	require.NoError(t, authdir.WriteUser("Main", token, seed))
	userFile, err := authdir.GetUserPath("main")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(configDir, "foojank", "users", "main"), userFile)
}

func TestUserRootOverride(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustUser(t, "main")
	require.NoError(t, authdir.WriteUser("Main", token, seed))
	userFile, err := authdir.GetUserPath("Main")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "users", "main"), userFile)

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, users)
}

func mustUser(t *testing.T, name string) (string, []byte) {
	t.Helper()

	key, err := auth.NewUserKey()
	require.NoError(t, err)

	claims, err := auth.NewUserJWT(name, jwt.Permissions{}, key)
	require.NoError(t, err)

	accountKey, err := auth.NewAccountKey()
	require.NoError(t, err)
	token, err := claims.Encode(accountKey)
	require.NoError(t, err)

	seed, err := key.Seed()
	require.NoError(t, err)
	return token, seed
}
