package authdir_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/authdir"
)

func TestCreateAccountRoundTrip(t *testing.T) {
	useTempHome(t)

	token, seed := mustAccount(t, "Main", "primary")
	require.NoError(t, authdir.CreateAccount("Main", token, seed))

	gotToken, gotSeed, err := authdir.ReadAccount("MAIN")
	require.NoError(t, err)
	require.Equal(t, token, gotToken)
	require.Equal(t, seed, gotSeed)

	claims, err := authdir.GetAccountJWT("main")
	require.NoError(t, err)
	require.Equal(t, "Main", claims.Name)
	require.Equal(t, "primary", claims.Description)
	require.Equal(t, claims.Subject, publicKey(t, seed))

	key, err := authdir.GetAccountKey("main")
	require.NoError(t, err)
	keySeed, err := key.Seed()
	require.NoError(t, err)
	require.Equal(t, seed, keySeed)

	pth := accountFile(t, "main")
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

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, accounts)
}

func TestCreateAccountExists(t *testing.T) {
	useTempHome(t)

	originalToken, originalSeed := mustAccount(t, "main", "original")
	replacementToken, replacementSeed := mustAccount(t, "main", "replacement")

	require.NoError(t, authdir.CreateAccount("Foo", originalToken, originalSeed))

	err := authdir.CreateAccount("foo", replacementToken, replacementSeed)
	require.ErrorIs(t, err, authdir.ErrAccountExists)
	require.NotErrorIs(t, err, os.ErrExist)

	gotToken, gotSeed, err := authdir.ReadAccount("foo")
	require.NoError(t, err)
	require.Equal(t, originalToken, gotToken)
	require.Equal(t, originalSeed, gotSeed)
}

func TestUpdateAccount(t *testing.T) {
	useTempHome(t)

	originalToken, originalSeed := mustAccount(t, "main", "original")
	replacementToken, replacementSeed := mustAccount(t, "main", "replacement")

	require.NoError(t, authdir.CreateAccount("main", originalToken, originalSeed))
	require.NoError(t, authdir.UpdateAccount("MAIN", replacementToken, replacementSeed))

	gotToken, gotSeed, err := authdir.ReadAccount("main")
	require.NoError(t, err)
	require.Equal(t, replacementToken, gotToken)
	require.Equal(t, replacementSeed, gotSeed)

	claims, err := authdir.GetAccountJWT("main")
	require.NoError(t, err)
	require.Equal(t, "replacement", claims.Description)
	require.Equal(t, claims.Subject, publicKey(t, replacementSeed))

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, accounts)
}

func TestUpdateAccountMissing(t *testing.T) {
	useTempHome(t)

	token, seed := mustAccount(t, "ghost", "missing")
	err := authdir.UpdateAccount("ghost", token, seed)
	require.ErrorIs(t, err, authdir.ErrAccountNotFound)
	require.NotErrorIs(t, err, os.ErrNotExist)

	_, _, err = authdir.ReadAccount("ghost")
	require.ErrorIs(t, err, authdir.ErrAccountNotFound)

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	require.Empty(t, accounts)
}

func TestRejectInvalidAccountCredentials(t *testing.T) {
	useTempHome(t)

	token, seed := mustAccount(t, "main", "")
	userToken, userSeed := mustUser(t, "user")
	operatorSeed := mustOperatorSeed(t)

	tests := []struct {
		name    string
		token   string
		seed    []byte
		wantErr string
	}{
		{name: "malformed jwt", token: "not-a-jwt", seed: seed, wantErr: "invalid account JWT"},
		{name: "user jwt", token: userToken, seed: seed, wantErr: "invalid account JWT"},
		{name: "empty jwt", token: "", seed: seed, wantErr: "invalid account JWT"},
		{name: "malformed seed", token: token, seed: []byte("nope"), wantErr: "invalid account seed"},
		{name: "prefix only seed", token: token, seed: []byte("SA"), wantErr: "invalid account seed"},
		{name: "user seed", token: token, seed: userSeed, wantErr: "invalid account seed"},
		{name: "operator seed", token: token, seed: operatorSeed, wantErr: "invalid account seed"},
		{name: "empty seed", token: token, seed: nil, wantErr: "invalid account seed"},
		{name: "jwt checked first", token: "not-a-jwt", seed: nil, wantErr: "invalid account JWT"},
	}

	writers := []struct {
		name string
		fn   func(string, string, []byte) error
	}{
		{name: "create", fn: authdir.CreateAccount},
		{name: "update", fn: authdir.UpdateAccount},
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

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	require.Empty(t, accounts)
}

func TestReadAccountMissing(t *testing.T) {
	useTempHome(t)

	_, _, err := authdir.ReadAccount("missing")
	require.ErrorIs(t, err, authdir.ErrAccountNotFound)

	_, err = authdir.GetAccountJWT("missing")
	require.ErrorIs(t, err, authdir.ErrAccountNotFound)

	_, err = authdir.GetAccountKey("missing")
	require.ErrorIs(t, err, authdir.ErrAccountNotFound)

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	require.Empty(t, accounts)
}

func TestReadAccountCorrupt(t *testing.T) {
	useTempHome(t)

	token, seed := mustAccount(t, "bad", "")
	require.NoError(t, authdir.CreateAccount("bad", token, seed))
	require.NoError(t, os.WriteFile(accountFile(t, "bad"), []byte("not a creds file"), 0o600))

	_, _, err := authdir.ReadAccount("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountNotFound)
	require.ErrorContains(t, err, "cannot decode JWT")

	_, err = authdir.GetAccountJWT("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountNotFound)

	_, err = authdir.GetAccountKey("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountNotFound)
}

func TestAccountRootOverride(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustAccount(t, "Main", "primary")
	require.NoError(t, authdir.CreateAccount("Main", token, seed))
	_, err := os.Stat(accountFile(t, "main"))
	require.NoError(t, err)

	gotToken, gotSeed, err := authdir.ReadAccount("MAIN")
	require.NoError(t, err)
	require.Equal(t, token, gotToken)
	require.Equal(t, seed, gotSeed)

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, accounts)
}

func TestAccountConfigDirUnavailable(t *testing.T) {
	setRoot(t, "")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustAccount(t, "main", "")
	require.Error(t, authdir.CreateAccount("main", token, seed))
	require.Error(t, authdir.UpdateAccount("main", token, seed))

	_, _, err := authdir.ReadAccount("main")
	require.Error(t, err)
	_, err = authdir.GetAccountJWT("main")
	require.Error(t, err)
	_, err = authdir.GetAccountKey("main")
	require.Error(t, err)
	_, err = authdir.ListAccounts()
	require.Error(t, err)
	require.Error(t, authdir.DeleteAccount("main"))
}

func TestCreateAccountParentIsFile(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "accounts"), []byte("x"), 0o600))

	token, seed := mustAccount(t, "main", "")
	err := authdir.CreateAccount("main", token, seed)
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountExists)
}

func TestUpdateAccountTargetIsDirectory(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "accounts", "main", "account"), 0o700))

	token, seed := mustAccount(t, "main", "")
	err := authdir.UpdateAccount("main", token, seed)
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountNotFound)
}

func TestReadAccountTargetIsDirectory(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "accounts", "main", "account"), 0o700))

	_, _, err := authdir.ReadAccount("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountNotFound)

	_, err = authdir.GetAccountJWT("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountNotFound)

	_, err = authdir.GetAccountKey("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrAccountNotFound)
}

func TestListAccountsStoreIsFile(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "accounts"), []byte("x"), 0o600))

	_, err := authdir.ListAccounts()
	require.Error(t, err)
}

func TestDeleteAccountPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission checks do not apply to root")
	}

	useTempHome(t)
	token, seed := mustAccount(t, "locked", "")
	require.NoError(t, authdir.CreateAccount("locked", token, seed))

	dir := filepath.Dir(accountFile(t, "locked"))
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})
	require.NoError(t, os.Chmod(dir, 0o500))

	err := authdir.DeleteAccount("locked")
	require.Error(t, err)
}

func TestReadAccountRejectsUndecodableSeed(t *testing.T) {
	useTempHome(t)

	token, _ := mustAccount(t, "main", "")
	writeCreds(t, accountFile(t, "main"), token, nil)

	_, _, err := authdir.ReadAccount("main")
	require.ErrorContains(t, err, "cannot decode decorated seed")
}

func TestReadAccountRejectsNonAccountSeed(t *testing.T) {
	useTempHome(t)

	token, _ := mustAccount(t, "main", "")
	_, userSeed := mustUser(t, "other")
	writeCreds(t, accountFile(t, "main"), token, userSeed)

	_, _, err := authdir.ReadAccount("main")
	require.EqualError(t, err, "invalid account seed")
}

func TestListAndDeleteAccounts(t *testing.T) {
	useTempHome(t)

	alphaToken, alphaSeed := mustAccount(t, "alpha", "")
	betaToken, betaSeed := mustAccount(t, "beta", "")
	require.NoError(t, authdir.CreateAccount("Alpha", alphaToken, alphaSeed))
	require.NoError(t, authdir.CreateAccount("beta", betaToken, betaSeed))

	accounts, err := authdir.ListAccounts()
	require.NoError(t, err)
	slices.Sort(accounts)
	require.Equal(t, []string{"alpha", "beta"}, accounts)

	require.NoError(t, authdir.DeleteAccount("ALPHA"))
	_, _, err = authdir.ReadAccount("alpha")
	require.ErrorIs(t, err, authdir.ErrAccountNotFound)

	_, _, err = authdir.ReadAccount("beta")
	require.NoError(t, err)

	require.NoError(t, authdir.DeleteAccount("alpha"))
	require.NoError(t, authdir.CreateAccount("alpha", alphaToken, alphaSeed))
	_, _, err = authdir.ReadAccount("alpha")
	require.NoError(t, err)
}

func useTempHome(t *testing.T) {
	t.Helper()

	setRoot(t, "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)
	rel, err := filepath.Rel(home, configDir)
	require.NoError(t, err)
	require.True(t, filepath.IsLocal(rel), "config dir %q escaped temp home %q", configDir, home)
}

func setRoot(t *testing.T, dir string) {
	t.Helper()

	prev := authdir.Root
	authdir.Root = dir
	t.Cleanup(func() {
		authdir.Root = prev
	})
}

func storeRoot(t *testing.T) string {
	t.Helper()

	if authdir.Root != "" {
		return authdir.Root
	}

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)
	return filepath.Join(configDir, "foojank")
}

func accountFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(storeRoot(t), "accounts", strings.ToLower(name), "account")
}

func writeCreds(t *testing.T, path, token string, seed []byte) {
	t.Helper()

	data, err := jwt.DecorateJWT(token)
	require.NoError(t, err)
	if seed != nil {
		decorated, err := jwt.DecorateSeed(seed)
		require.NoError(t, err)
		data = append(data, decorated...)
	}

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func mustAccount(t *testing.T, name, description string) (string, []byte) {
	t.Helper()

	key, err := auth.NewAccountKey()
	require.NoError(t, err)

	claims, err := auth.NewAccountJWT(name, key)
	require.NoError(t, err)
	claims.Description = description

	token, err := claims.Encode(key)
	require.NoError(t, err)

	seed, err := key.Seed()
	require.NoError(t, err)
	return token, seed
}

func mustOperatorSeed(t *testing.T) []byte {
	t.Helper()

	key, err := nkeys.CreateOperator()
	require.NoError(t, err)
	seed, err := key.Seed()
	require.NoError(t, err)
	return seed
}

func publicKey(t *testing.T, seed []byte) string {
	t.Helper()

	key, err := nkeys.FromSeed(seed)
	require.NoError(t, err)
	pub, err := key.PublicKey()
	require.NoError(t, err)
	return pub
}
