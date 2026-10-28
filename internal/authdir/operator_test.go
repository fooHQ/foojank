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

func TestCreateOperatorRoundTrip(t *testing.T) {
	useTempHome(t)

	token, seed := mustOperator(t, "Main")
	require.NoError(t, authdir.CreateOperator("Main", token, seed))

	gotToken, gotSeed, err := authdir.ReadOperator("MAIN")
	require.NoError(t, err)
	require.Equal(t, token, gotToken)
	require.Equal(t, seed, gotSeed)

	claims, err := authdir.GetOperatorJWT("main")
	require.NoError(t, err)
	require.Equal(t, "Main", claims.Name)
	require.Equal(t, claims.Subject, publicKey(t, seed))

	key, err := authdir.GetOperatorKey("main")
	require.NoError(t, err)
	keySeed, err := key.Seed()
	require.NoError(t, err)
	require.Equal(t, seed, keySeed)

	pth, err := authdir.GetOperatorPath("main")
	require.NoError(t, err)
	require.Equal(t, operatorFile(t, "main"), pth)

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

	operators, err := authdir.ListOperators()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, operators)
}

func TestCreateOperatorExists(t *testing.T) {
	useTempHome(t)

	originalToken, originalSeed := mustOperator(t, "original")
	replacementToken, replacementSeed := mustOperator(t, "replacement")

	require.NoError(t, authdir.CreateOperator("Foo", originalToken, originalSeed))

	err := authdir.CreateOperator("foo", replacementToken, replacementSeed)
	require.ErrorIs(t, err, authdir.ErrOperatorExists)
	require.NotErrorIs(t, err, os.ErrExist)

	gotToken, gotSeed, err := authdir.ReadOperator("foo")
	require.NoError(t, err)
	require.Equal(t, originalToken, gotToken)
	require.Equal(t, originalSeed, gotSeed)
}

func TestUpdateOperator(t *testing.T) {
	useTempHome(t)

	originalToken, originalSeed := mustOperator(t, "original")
	replacementToken, replacementSeed := mustOperator(t, "replacement")

	require.NoError(t, authdir.CreateOperator("main", originalToken, originalSeed))
	require.NoError(t, authdir.UpdateOperator("MAIN", replacementToken, replacementSeed))

	gotToken, gotSeed, err := authdir.ReadOperator("main")
	require.NoError(t, err)
	require.Equal(t, replacementToken, gotToken)
	require.Equal(t, replacementSeed, gotSeed)

	claims, err := authdir.GetOperatorJWT("main")
	require.NoError(t, err)
	require.Equal(t, "replacement", claims.Name)
	require.Equal(t, claims.Subject, publicKey(t, replacementSeed))

	operators, err := authdir.ListOperators()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, operators)
}

func TestUpdateOperatorMissing(t *testing.T) {
	useTempHome(t)

	token, seed := mustOperator(t, "ghost")
	err := authdir.UpdateOperator("ghost", token, seed)
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)
	require.NotErrorIs(t, err, os.ErrNotExist)

	_, _, err = authdir.ReadOperator("ghost")
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)

	operators, err := authdir.ListOperators()
	require.NoError(t, err)
	require.Empty(t, operators)
}

func TestRejectInvalidOperatorCredentials(t *testing.T) {
	useTempHome(t)

	token, seed := mustOperator(t, "main")
	accountToken, accountSeed := mustAccount(t, "main", "")
	userToken, userSeed := mustUser(t, "user")

	tests := []struct {
		name    string
		token   string
		seed    []byte
		wantErr string
	}{
		{name: "malformed jwt", token: "not-a-jwt", seed: seed, wantErr: "invalid operator JWT"},
		{name: "account jwt", token: accountToken, seed: seed, wantErr: "invalid operator JWT"},
		{name: "user jwt", token: userToken, seed: seed, wantErr: "invalid operator JWT"},
		{name: "empty jwt", token: "", seed: seed, wantErr: "invalid operator JWT"},
		{name: "malformed seed", token: token, seed: []byte("nope"), wantErr: "invalid operator seed"},
		{name: "prefix only seed", token: token, seed: []byte("SO"), wantErr: "invalid operator seed"},
		{name: "account seed", token: token, seed: accountSeed, wantErr: "invalid operator seed"},
		{name: "user seed", token: token, seed: userSeed, wantErr: "invalid operator seed"},
		{name: "empty seed", token: token, seed: nil, wantErr: "invalid operator seed"},
		{name: "jwt checked first", token: "not-a-jwt", seed: nil, wantErr: "invalid operator JWT"},
	}

	writers := []struct {
		name string
		fn   func(string, string, []byte) error
	}{
		{name: "create", fn: authdir.CreateOperator},
		{name: "update", fn: authdir.UpdateOperator},
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

	operators, err := authdir.ListOperators()
	require.NoError(t, err)
	require.Empty(t, operators)
}

func TestReadOperatorMissing(t *testing.T) {
	useTempHome(t)

	_, _, err := authdir.ReadOperator("missing")
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)

	_, err = authdir.GetOperatorJWT("missing")
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)

	_, err = authdir.GetOperatorKey("missing")
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)

	_, err = authdir.GetOperatorPath("missing")
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)

	operators, err := authdir.ListOperators()
	require.NoError(t, err)
	require.Empty(t, operators)
}

func TestReadOperatorCorrupt(t *testing.T) {
	useTempHome(t)

	token, seed := mustOperator(t, "bad")
	require.NoError(t, authdir.CreateOperator("bad", token, seed))

	pth, err := authdir.GetOperatorPath("bad")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(pth, []byte("not a creds file"), 0o600))

	_, _, err = authdir.ReadOperator("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)
	require.ErrorContains(t, err, "cannot decode JWT")

	_, err = authdir.GetOperatorJWT("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)

	_, err = authdir.GetOperatorKey("bad")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)
}

func TestOperatorRootOverride(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustOperator(t, "Main")
	require.NoError(t, authdir.CreateOperator("Main", token, seed))

	operatorPath, err := authdir.GetOperatorPath("MAIN")
	require.NoError(t, err)
	require.Equal(t, operatorFile(t, "main"), operatorPath)

	gotToken, gotSeed, err := authdir.ReadOperator("main")
	require.NoError(t, err)
	require.Equal(t, token, gotToken)
	require.Equal(t, seed, gotSeed)

	operators, err := authdir.ListOperators()
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, operators)
}

func TestOperatorConfigDirUnavailable(t *testing.T) {
	setRoot(t, "")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	token, seed := mustOperator(t, "main")
	require.Error(t, authdir.CreateOperator("main", token, seed))
	require.Error(t, authdir.UpdateOperator("main", token, seed))

	_, _, err := authdir.ReadOperator("main")
	require.Error(t, err)
	_, err = authdir.GetOperatorJWT("main")
	require.Error(t, err)
	_, err = authdir.GetOperatorKey("main")
	require.Error(t, err)
	_, err = authdir.GetOperatorPath("main")
	require.Error(t, err)
	_, err = authdir.ListOperators()
	require.Error(t, err)
	require.Error(t, authdir.DeleteOperator("main"))
}

func TestCreateOperatorParentIsFile(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "operators"), []byte("x"), 0o600))

	token, seed := mustOperator(t, "main")
	err := authdir.CreateOperator("main", token, seed)
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorExists)
}

func TestUpdateOperatorTargetIsDirectory(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "operators", "main"), 0o700))

	token, seed := mustOperator(t, "main")
	err := authdir.UpdateOperator("main", token, seed)
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)
}

func TestReadOperatorTargetIsDirectory(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "operators", "main"), 0o700))

	_, _, err := authdir.ReadOperator("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)

	_, err = authdir.GetOperatorJWT("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)

	_, err = authdir.GetOperatorKey("main")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)
}

func TestListOperatorsStoreIsFile(t *testing.T) {
	dir := t.TempDir()
	setRoot(t, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "operators"), []byte("x"), 0o600))

	_, err := authdir.ListOperators()
	require.Error(t, err)
}

func TestGetOperatorPathPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission checks do not apply to root")
	}

	useTempHome(t)
	token, seed := mustOperator(t, "hidden")
	require.NoError(t, authdir.CreateOperator("hidden", token, seed))

	dir := filepath.Dir(operatorFile(t, "hidden"))
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})
	require.NoError(t, os.Chmod(dir, 0))

	_, err := authdir.GetOperatorPath("hidden")
	require.Error(t, err)
	require.NotErrorIs(t, err, authdir.ErrOperatorNotFound)
}

func TestDeleteOperatorPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission checks do not apply to root")
	}

	useTempHome(t)
	token, seed := mustOperator(t, "locked")
	require.NoError(t, authdir.CreateOperator("locked", token, seed))

	dir := filepath.Dir(operatorFile(t, "locked"))
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})
	require.NoError(t, os.Chmod(dir, 0o500))

	err := authdir.DeleteOperator("locked")
	require.Error(t, err)
}

func TestReadOperatorRejectsUndecodableSeed(t *testing.T) {
	useTempHome(t)

	token, _ := mustOperator(t, "main")
	writeCreds(t, operatorFile(t, "main"), token, nil)

	_, _, err := authdir.ReadOperator("main")
	require.ErrorContains(t, err, "cannot decode decorated seed")
}

func TestReadOperatorRejectsNonOperatorSeed(t *testing.T) {
	useTempHome(t)

	token, _ := mustOperator(t, "main")
	_, accountSeed := mustAccount(t, "other", "")
	writeCreds(t, operatorFile(t, "main"), token, accountSeed)

	_, _, err := authdir.ReadOperator("main")
	require.EqualError(t, err, "invalid operator seed")
}

func TestListAndDeleteOperators(t *testing.T) {
	useTempHome(t)

	alphaToken, alphaSeed := mustOperator(t, "alpha")
	betaToken, betaSeed := mustOperator(t, "beta")
	require.NoError(t, authdir.CreateOperator("Alpha", alphaToken, alphaSeed))
	require.NoError(t, authdir.CreateOperator("beta", betaToken, betaSeed))

	operators, err := authdir.ListOperators()
	require.NoError(t, err)
	slices.Sort(operators)
	require.Equal(t, []string{"alpha", "beta"}, operators)

	require.NoError(t, authdir.DeleteOperator("ALPHA"))
	_, _, err = authdir.ReadOperator("alpha")
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)

	_, _, err = authdir.ReadOperator("beta")
	require.NoError(t, err)

	require.NoError(t, authdir.DeleteOperator("alpha"))
	require.NoError(t, authdir.CreateOperator("alpha", alphaToken, alphaSeed))
	_, _, err = authdir.ReadOperator("alpha")
	require.NoError(t, err)
}

func operatorFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(storeRoot(t), "operators", strings.ToLower(name))
}

func mustOperator(t *testing.T, name string) (string, []byte) {
	t.Helper()

	key, err := auth.NewOperatorKey()
	require.NoError(t, err)

	claims, err := auth.NewOperatorJWT(name, key)
	require.NoError(t, err)

	token, err := claims.Encode(key)
	require.NoError(t, err)

	seed, err := key.Seed()
	require.NoError(t, err)
	return token, seed
}
