package authdir

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountExists   = errors.New("account already exists")
)

func CreateAccount(name string, accountJWT string, accountSeed []byte) error {
	if !isAccountJWT(accountJWT) {
		return errors.New("invalid account JWT")
	}

	if !isAccountSeed(accountSeed) {
		return errors.New("invalid account seed")
	}

	data, err := decorate(accountJWT, accountSeed)
	if err != nil {
		return err
	}

	err = writeAccountData(name, data, os.O_CREATE|os.O_WRONLY|os.O_EXCL)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrAccountExists
		}
		return err
	}

	return nil
}

func UpdateAccount(name string, accountJWT string, accountSeed []byte) error {
	if !isAccountJWT(accountJWT) {
		return errors.New("invalid account JWT")
	}

	if !isAccountSeed(accountSeed) {
		return errors.New("invalid account seed")
	}

	data, err := decorate(accountJWT, accountSeed)
	if err != nil {
		return err
	}

	err = writeAccountData(name, data, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrAccountNotFound
		}
		return err
	}

	return nil
}

func GetAccountKey(name string) (nkeys.KeyPair, error) {
	data, err := readAccountData(name)
	if err != nil {
		return nil, err
	}
	return jwt.ParseDecoratedNKey(data)
}

func GetAccountJWT(name string) (*jwt.AccountClaims, error) {
	data, err := readAccountData(name)
	if err != nil {
		return nil, err
	}

	accountJWT, err := jwt.ParseDecoratedJWT(data)
	if err != nil {
		return nil, err
	}

	return jwt.DecodeAccountClaims(accountJWT)
}

func ReadAccount(name string) (string, []byte, error) {
	data, err := readAccountData(name)
	if err != nil {
		return "", nil, err
	}

	accountJWT, err := jwt.ParseDecoratedJWT(data)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode decorated JWT: %w", err)
	}

	// Validate that the JWT is account JWT.
	_, err = jwt.DecodeAccountClaims(accountJWT)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode JWT: %w", err)
	}

	account, err := jwt.ParseDecoratedNKey(data)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode decorated seed: %w", err)
	}

	accountSeed, err := account.Seed()
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode seed: %w", err)
	}

	if !isAccountSeed(accountSeed) {
		return "", nil, errors.New("invalid account seed")
	}

	return accountJWT, accountSeed, nil
}

func ListAccounts() ([]string, error) {
	pth, err := accountRootPath()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var accounts []string
	for _, file := range files {
		accounts = append(accounts, file.Name())
	}

	return accounts, nil
}

func DeleteAccount(name string) error {
	pth, err := accountPath(name)
	if err != nil {
		return err
	}

	err = os.RemoveAll(pth)
	if err != nil {
		return err
	}
	return nil
}

func accountRootPath() (string, error) {
	root, err := rootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "accounts"), nil
}

func accountPath(name string) (string, error) {
	name = strings.ToLower(name)
	root, err := accountRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, name, "account"), nil
}

func decorate(accountJWT string, accountSeed []byte) ([]byte, error) {
	jwtDecorated, err := jwt.DecorateJWT(accountJWT)
	if err != nil {
		return nil, fmt.Errorf("cannot decorate JWT: %w", err)
	}

	seedDecorated, err := jwt.DecorateSeed(accountSeed)
	if err != nil {
		return nil, fmt.Errorf("cannot decorate seed: %w", err)
	}

	return bytes.Join([][]byte{jwtDecorated, seedDecorated}, []byte("")), nil
}

func isAccountJWT(s string) bool {
	_, err := jwt.DecodeAccountClaims(s)
	return err == nil
}

func isAccountSeed(key []byte) bool {
	kp, err := nkeys.FromSeed(key)
	if err != nil {
		return false
	}

	pub, err := kp.PublicKey()
	if err != nil {
		return false
	}

	return nkeys.IsValidPublicAccountKey(pub)
}

func writeAccountData(name string, data []byte, flag int) error {
	pth, err := accountPath(name)
	if err != nil {
		return err
	}

	// Create parent directories only when the caller is allowed to create the
	// file. Update opens an existing file; creating the parent first would
	// leave an empty directory that ListAccounts treats as an account.
	if flag&os.O_CREATE != 0 {
		err = os.MkdirAll(filepath.Dir(pth), 0o700)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(pth, flag, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func readAccountData(name string) ([]byte, error) {
	pth, err := accountPath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	return data, nil
}
