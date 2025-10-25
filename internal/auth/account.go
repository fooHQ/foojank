package auth

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

const (
	accountRootPath = "foojank/accounts"
	accountPathT    = accountRootPath + "/%s/account"
)

func WriteAccount(name string, accountJWT string, accountSeed []byte) error {
	// Validate that the JWT is an account JWT.
	_, err := jwt.DecodeAccountClaims(accountJWT)
	if err != nil {
		return fmt.Errorf("cannot decode JWT: %w", err)
	}

	if !isAccountSeed(accountSeed) {
		return errors.New("invalid account seed")
	}

	jwtDecorated, err := jwt.DecorateJWT(accountJWT)
	if err != nil {
		return fmt.Errorf("cannot encode decorated JWT: %w", err)
	}

	seedDecorated, err := jwt.DecorateSeed(accountSeed)
	if err != nil {
		return fmt.Errorf("cannot encode decorated seed: %w", err)
	}

	data := bytes.Join([][]byte{jwtDecorated, seedDecorated}, []byte(""))

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	pth := filepath.Join(configDir, fmt.Sprintf(accountPathT, name))

	err = os.MkdirAll(filepath.Dir(pth), 0700)
	if err != nil {
		return err
	}

	err = os.WriteFile(pth, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func ReadAccount(name string) (string, []byte, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", nil, err
	}

	pth := filepath.Join(configDir, fmt.Sprintf(accountPathT, name))

	data, err := os.ReadFile(pth)
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

func NewAccount(name string) (string, []byte, error) {
	account, err := nkeys.CreateAccount()
	if err != nil {
		return "", nil, fmt.Errorf("cannot generate key pair: %w", err)
	}

	publicKey, err := account.PublicKey()
	if err != nil {
		return "", nil, fmt.Errorf("cannot get public key: %w", err)
	}

	claims := jwt.NewAccountClaims(publicKey)
	claims.Name = name
	claims.Limits.JetStreamLimits = jwt.JetStreamLimits{
		DiskStorage:   -1,
		MemoryStorage: -1,
	}

	accountJWT, err := claims.Encode(account)
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode JWT: %w", err)
	}

	accountSeed, err := account.Seed()
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode seed: %w", err)
	}

	return accountJWT, accountSeed, nil
}

func isAccountSeed(key []byte) bool {
	return bytes.HasPrefix(key, []byte("SA"))
}
