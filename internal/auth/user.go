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
	userPathT = accountRootPath + "/%s/user"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func WriteUser(name string, userJWT string, userSeed []byte) error {
	// Validate that the JWT is a user JWT.
	_, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return fmt.Errorf("cannot decode JWT: %w", err)
	}

	if !isUserSeed(userSeed) {
		return errors.New("invalid user seed")
	}

	jwtDecorated, err := jwt.DecorateJWT(userJWT)
	if err != nil {
		return fmt.Errorf("cannot encode decorated JWT: %w", err)
	}

	seedDecorated, err := jwt.DecorateSeed(userSeed)
	if err != nil {
		return fmt.Errorf("cannot encode decorated seed: %w", err)
	}

	data := bytes.Join([][]byte{jwtDecorated, seedDecorated}, []byte(""))

	pth, err := UserPath(name)
	if err != nil {
		return err
	}

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

func ReadUser(name string) (string, []byte, error) {
	pth, err := UserPath(name)
	if err != nil {
		return "", nil, err
	}

	data, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, ErrUserNotFound
		}
		return "", nil, err
	}

	userJWT, err := jwt.ParseDecoratedJWT(data)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode decorated JWT: %w", err)
	}

	// Validate that the JWT is a user JWT.
	_, err = jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode JWT: %w", err)
	}

	user, err := jwt.ParseDecoratedNKey(data)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode decorated seed: %w", err)
	}

	userSeed, err := user.Seed()
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode seed: %w", err)
	}

	if !isUserSeed(userSeed) {
		return "", nil, errors.New("invalid user seed")
	}

	return userJWT, userSeed, nil
}

func NewUser(name string, accountSeed []byte, perms jwt.Permissions) (string, []byte, error) {
	account, err := nkeys.FromSeed(accountSeed)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode account seed: %w", err)
	}

	accountPublicKey, err := account.PublicKey()
	if err != nil {
		return "", nil, fmt.Errorf("cannot get account public key: %w", err)
	}

	user, err := nkeys.CreateUser()
	if err != nil {
		return "", nil, fmt.Errorf("cannot generate key pair: %w", err)
	}

	userPublicKey, err := user.PublicKey()
	if err != nil {
		return "", nil, fmt.Errorf("cannot get public key: %w", err)
	}

	claims := jwt.NewUserClaims(userPublicKey)
	claims.Name = name
	claims.IssuerAccount = accountPublicKey
	claims.Permissions = perms

	userJWT, err := claims.Encode(account)
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode JWT: %w", err)
	}

	userSeed, err := user.Seed()
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode seed: %w", err)
	}

	return userJWT, userSeed, nil
}

func UserPath(name string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, fmt.Sprintf(userPathT, filepath.Clean(name))), nil
}

func isUserSeed(key []byte) bool {
	return bytes.HasPrefix(key, []byte("SU"))
}
