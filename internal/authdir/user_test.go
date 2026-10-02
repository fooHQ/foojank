package authdir_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nats-io/jwt/v2"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/authdir"
)

func TestCreateUserRoundTrip(t *testing.T) {
	useTempHome(t)

	token, seed := mustUser(t, "Main")
	require.NoError(t, authdir.CreateUser("Main", token, seed))

	gotToken, gotSeed, err := authdir.ReadUser("MAIN")
	require.NoError(t, err)
	require.Equal(t, token, gotToken)
	require.Equal(t, seed, gotSeed)

	claims, err := authdir.GetUserJWT("main")
	require.NoError(t, err)
	require.Equal(t, "Main", claims.Name)
	require.Equal(t, claims.Subject, publicKey(t, seed))

	key, err := authdir.GetUserKey("main")
	require.NoError(t, err)
	keySeed, err := key.Seed()
	require.NoError(t, err)
	require.Equal(t, seed, keySeed)

	pth, err := authdir.GetUserPath("main")
	require.NoError(t, err)
	require.Equal(t, userFile(t, "main"), pth)

	data, err := os.ReadFile(pth)
	require.NoError(t, err)
	jwtDecorated, err := jwt.DecorateJWT(token)
	require.NoError(t, err)
	seedDecorated, err := jwt.DecorateSeed(seed)
	require.NoError(t, err)
	require.Equal(t, append(jwtDecorated, seedDecorated...), data)

	info, err := os.Stat(pth)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	dirInfo, err := os.Stat(filepath.Dir(pth))
	require.NoError(t, err)
	require.True(t, dirInfo.IsDir())
	require.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, users)
}

func TestCreateUserExists(t *testing.T) {
	useTempHome(t)

	originalToken, originalSeed := mustUser(t, "original")
	replacementToken, replacementSeed := mustUser(t, "replacement")

	require.NoError(t, authdir.CreateUser("Foo", originalToken, originalSeed))

	err := authdir.CreateUser("foo", replacementToken, replacementSeed)
	require.ErrorIs(t, err, authdir.ErrUserExists)
	require.NotErrorIs(t, err, os.ErrExist)

	gotToken, gotSeed, err := authdir.ReadUser("foo")
	require.NoError(t, err)
	require.Equal(t, originalToken, gotToken)
	require.Equal(t, originalSeed, gotSeed)
}

func TestUpdateUser(t *testing.T) {
	useTempHome(t)

	originalToken, originalSeed := mustUser(t, "original")
	replacementToken, replacementSeed := mustUser(t, "replacement")

	require.NoError(t, authdir.CreateUser("main", originalToken, originalSeed))
	require.NoError(t, authdir.UpdateUser("MAIN", replacementToken, replacementSeed))

	gotToken, gotSeed, err := authdir.ReadUser("main")
	require.NoError(t, err)
	require.Equal(t, replacementToken, gotToken)
	require.Equal(t, replacementSeed, gotSeed)

	claims, err := authdir.GetUserJWT("main")
	require.NoError(t, err)
	require.Equal(t, "replacement", claims.Name)
	require.Equal(t, claims.Subject, publicKey(t, replacementSeed))

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, users)
}

func TestUpdateUserMissing(t *testing.T) {
	useTempHome(t)

	token, seed := mustUser(t, "ghost")
	err := authdir.UpdateUser("ghost", token, seed)
	require.ErrorIs(t, err, authdir.ErrUserNotFound)
	require.NotErrorIs(t, err, os.ErrNotExist)

	_, _, err = authdir.ReadUser("ghost")
	require.ErrorIs(t, err, authdir.ErrUserNotFound)

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	require.Empty(t, users)
}

func TestRejectInvalidUserCredentials(t *testing.T) {
	useTempHome(t)

	token, seed := mustUser(t, "main")
	accountToken, accountSeed := mustAccount(t, "main", "")
	operatorSeed := mustOperatorSeed(t)

	tests := []struct {
		name    string
		token   string
		seed    []byte
		wantErr string
	}{
		{name: "malformed jwt", token: "not-a-jwt", seed: seed, wantErr: "invalid user JWT"},
		{name: "account jwt", token: accountToken, seed: seed, wantErr: "invalid user JWT"},
		{name: "empty jwt", token: "", seed: seed, wantErr: "invalid user JWT"},
		{name: "malformed seed", token: token, seed: []byte("nope"), wantErr: "invalid user seed"},
		{name: "prefix only seed", token: token, seed: []byte("SU"), wantErr: "invalid user seed"},
		{name: "account seed", token: token, seed: accountSeed, wantErr: "invalid user seed"},
		{name: "operator seed", token: token, seed: operatorSeed, wantErr: "invalid user seed"},
		{name: "empty seed", token: token, seed: nil, wantErr: "invalid user seed"},
		{name: "jwt checked first", token: "not-a-jwt", seed: nil, wantErr: "invalid user JWT"},
	}

	writers := []struct {
		name string
		fn   func(string, string, []byte) error
	}{
		{name: "create", fn: authdir.CreateUser},
		{name: "update", fn: authdir.UpdateUser},
	}

	for _, writer := range writers {
		t.Run(writer.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					err := writer.fn("rejected", tt.token, tt.seed)
					require.EqualError(t, err, tt.wantErr)
				})
			}
		})
	}

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	require.Empty(t, users)
}

func TestReadUserMissing(t *testing.T) {
	useTempHome(t)

	_, _, err := authdir.ReadUser("missing")
	require.ErrorIs(t, err, authdir.ErrUserNotFound)

	_, err = authdir.GetUserJWT("missing")
	require.ErrorIs(t, err, authdir.ErrUserNotFound)

	_, err = authdir.GetUserKey("missing")
	require.ErrorIs(t, err, authdir.ErrUserNotFound)

	_, err = authdir.GetUserPath("missing")
	require.ErrorIs(t, err, authdir.ErrUserNotFound)

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	require.Empty(t, users)
}

func TestReadUserCorrupt(t *testing.T) {
	useTempHome(t)

	token, seed := mustUser(t, "bad")
	require.NoError(t, authdir.CreateUser("bad", token, seed))

	pth, err := authdir.GetUserPath("bad")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(pth, []byte("not a creds file"), 0o600))

	_, _, err = authdir.ReadUser("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)
	require.ErrorContains(t, err, "cannot decode JWT")

	_, err = authdir.GetUserJWT("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)

	_, err = authdir.GetUserKey("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)
}

func TestUserRootOverride(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustUser(t, "Main")
	require.NoError(t, authdir.CreateUser("Main", token, seed))

	userPath, err := authdir.GetUserPath("MAIN")
	require.NoError(t, err)
	require.Equal(t, userFile(t, "main"), userPath)

	gotToken, gotSeed, err := authdir.ReadUser("main")
	require.NoError(t, err)
	require.Equal(t, token, gotToken)
	require.Equal(t, seed, gotSeed)

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, users)
}

func TestUserConfigDirUnavailable(t *testing.T) {
	setRoot(t, "")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustUser(t, "main")
	require.Error(t, authdir.CreateUser("main", token, seed))
	require.Error(t, authdir.UpdateUser("main", token, seed))

	_, _, err := authdir.ReadUser("main")
	require.Error(t, err)
	_, err = authdir.GetUserJWT("main")
	require.Error(t, err)
	_, err = authdir.GetUserKey("main")
	require.Error(t, err)
	_, err = authdir.GetUserPath("main")
	require.Error(t, err)
	_, err = authdir.ListUsers()
	require.Error(t, err)
}

func TestCreateUserParentIsFile(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users"), []byte("x"), 0o600))

	token, seed := mustUser(t, "main")
	err := authdir.CreateUser("main", token, seed)
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserExists)
}

func TestUpdateUserTargetIsDirectory(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "users", "main"), 0o700))

	token, seed := mustUser(t, "main")
	err := authdir.UpdateUser("main", token, seed)
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)
}

func TestReadUserTargetIsDirectory(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "users", "main"), 0o700))

	_, _, err := authdir.ReadUser("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)

	_, err = authdir.GetUserJWT("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)

	_, err = authdir.GetUserKey("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)
}

func TestListUsersStoreIsFile(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users"), []byte("x"), 0o600))

	_, err := authdir.ListUsers()
	require.Error(t, err)
}

func TestGetUserPathPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission checks do not apply to root")
	}

	useTempHome(t)
	token, seed := mustUser(t, "hidden")
	require.NoError(t, authdir.CreateUser("hidden", token, seed))

	dir := filepath.Dir(userFile(t, "hidden"))
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})
	require.NoError(t, os.Chmod(dir, 0))

	_, err := authdir.GetUserPath("hidden")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrUserNotFound)
}

func TestReadUserRejectsUndecodableSeed(t *testing.T) {
	useTempHome(t)

	token, _ := mustUser(t, "main")
	writeCreds(t, userFile(t, "main"), token, nil)

	_, _, err := authdir.ReadUser("main")
	require.ErrorContains(t, err, "cannot decode decorated seed")
}

func TestReadUserRejectsNonUserSeed(t *testing.T) {
	useTempHome(t)

	token, _ := mustUser(t, "main")
	_, accountSeed := mustAccount(t, "other", "")
	writeCreds(t, userFile(t, "main"), token, accountSeed)

	_, _, err := authdir.ReadUser("main")
	require.EqualError(t, err, "invalid user seed")
}

func TestListUsers(t *testing.T) {
	useTempHome(t)

	alphaToken, alphaSeed := mustUser(t, "alpha")
	betaToken, betaSeed := mustUser(t, "beta")
	require.NoError(t, authdir.CreateUser("Alpha", alphaToken, alphaSeed))
	require.NoError(t, authdir.CreateUser("beta", betaToken, betaSeed))

	users, err := authdir.ListUsers()
	require.NoError(t, err)
	slices.Sort(users)
	require.Equal(t, []string{"alpha", "beta"}, users)
}

func userFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(storeRoot(t), "users", strings.ToLower(name))
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
