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

	got, err := authdir.GetOperator("MAIN")
	require.NoError(t, err)
	require.Equal(t, token, got.JWT())
	require.Equal(t, "Main", got.Name())
	require.Equal(t, "Main", got.Claims().Name)
	require.Equal(t, got.PublicKey(), publicKey(t, seed))
	gotSeed, err := got.KeyPair().Seed()
	require.NoError(t, err)
	require.Equal(t, seed, gotSeed)

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

	got, err := authdir.GetOperator("foo")
	require.NoError(t, err)
	require.Equal(t, originalToken, got.JWT())
	require.Equal(t, "original", got.Name())
	gotSeed, err := got.KeyPair().Seed()
	require.NoError(t, err)
	require.Equal(t, originalSeed, gotSeed)
}

func TestUpdateOperator(t *testing.T) {
	useTempHome(t)

	originalToken, originalSeed := mustOperator(t, "original")
	replacementToken, replacementSeed := mustOperator(t, "replacement")

	require.NoError(t, authdir.CreateOperator("main", originalToken, originalSeed))
	require.NoError(t, authdir.UpdateOperator("MAIN", replacementToken, replacementSeed))

	got, err := authdir.GetOperator("main")
	require.NoError(t, err)
	require.Equal(t, replacementToken, got.JWT())
	require.Equal(t, "replacement", got.Name())
	require.Equal(t, got.Claims().Subject, publicKey(t, replacementSeed))
	gotSeed, err := got.KeyPair().Seed()
	require.NoError(t, err)
	require.Equal(t, replacementSeed, gotSeed)

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

	_, err = authdir.GetOperator("ghost")
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

	_, err := authdir.GetOperator("missing")
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

	_, err = authdir.GetOperator("bad")
	require.EqualError(t, err, "no nkey seed found")
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

	got, err := authdir.GetOperator("main")
	require.NoError(t, err)
	require.Equal(t, token, got.JWT())
	gotSeed, err := got.KeyPair().Seed()
	require.NoError(t, err)
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

	_, err := authdir.GetOperator("main")
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

	_, err := authdir.GetOperator("main")
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

	_, err := authdir.GetOperator("main")
	require.EqualError(t, err, "no nkey seed found")
}

func TestReadOperatorRejectsNonOperatorSeed(t *testing.T) {
	useTempHome(t)

	token, _ := mustOperator(t, "main")
	_, accountSeed := mustAccount(t, "other", "")
	writeCreds(t, operatorFile(t, "main"), token, accountSeed)

	_, err := authdir.GetOperator("main")
	require.EqualError(t, err, "JWT and key mismatch")
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
	_, err = authdir.GetOperator("alpha")
	require.ErrorIs(t, err, authdir.ErrOperatorNotFound)

	got, err := authdir.GetOperator("beta")
	require.NoError(t, err)
	require.Equal(t, betaToken, got.JWT())

	require.NoError(t, authdir.DeleteOperator("alpha"))
	require.NoError(t, authdir.CreateOperator("alpha", alphaToken, alphaSeed))
	got, err = authdir.GetOperator("alpha")
	require.NoError(t, err)
	require.Equal(t, alphaToken, got.JWT())
	gotSeed, err := got.KeyPair().Seed()
	require.NoError(t, err)
	require.Equal(t, alphaSeed, gotSeed)
}

func operatorFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(storeRoot(t), "operators", strings.ToLower(name))
}

func mustOperator(t *testing.T, name string) (string, []byte) {
	t.Helper()

	operator, err := auth.GenerateOperator(name)
	require.NoError(t, err)

	token, err := operator.Encode(operator.KeyPair())
	require.NoError(t, err)

	seed, err := operator.KeyPair().Seed()
	require.NoError(t, err)
	return token, seed
}
