package authdir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

func CreateUser(name string, userJWT string, userSeed []byte) error {
	if !isUserJWT(userJWT) {
		return errors.New("invalid user JWT")
	}

	if !isUserSeed(userSeed) {
		return errors.New("invalid user seed")
	}

	data, err := decorate(userJWT, userSeed)
	if err != nil {
		return err
	}

	err = writeUserData(name, data, os.O_CREATE|os.O_WRONLY|os.O_EXCL)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrUserExists
		}
		return err
	}

	return nil
}

func UpdateUser(name string, userJWT string, userSeed []byte) error {
	if !isUserJWT(userJWT) {
		return errors.New("invalid user JWT")
	}

	if !isUserSeed(userSeed) {
		return errors.New("invalid user seed")
	}

	data, err := decorate(userJWT, userSeed)
	if err != nil {
		return err
	}

	err = writeUserData(name, data, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func GetUserKey(name string) (nkeys.KeyPair, error) {
	data, err := readUserData(name)
	if err != nil {
		return nil, err
	}
	return jwt.ParseDecoratedNKey(data)
}

func GetUserJWT(name string) (*jwt.UserClaims, error) {
	data, err := readUserData(name)
	if err != nil {
		return nil, err
	}

	userJWT, err := jwt.ParseDecoratedJWT(data)
	if err != nil {
		return nil, err
	}

	return jwt.DecodeUserClaims(userJWT)
}

func ReadUser(name string) (string, []byte, error) {
	data, err := readUserData(name)
	if err != nil {
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

func ListUsers() ([]string, error) {
	pth, err := userRootPath()
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

	var users []string
	for _, file := range files {
		users = append(users, file.Name())
	}

	return users, nil
}

func GetUserPath(name string) (string, error) {
	pth, err := userPath(name)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return pth, nil
}

func userRootPath() (string, error) {
	root, err := rootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "users"), nil
}

func userPath(name string) (string, error) {
	name = strings.ToLower(name)
	root, err := userRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, name), nil
}

func isUserJWT(s string) bool {
	_, err := jwt.DecodeUserClaims(s)
	return err == nil
}

func isUserSeed(key []byte) bool {
	kp, err := nkeys.FromSeed(key)
	if err != nil {
		return false
	}

	pub, err := kp.PublicKey()
	if err != nil {
		return false
	}

	return nkeys.IsValidPublicUserKey(pub)
}

func writeUserData(name string, data []byte, flag int) error {
	pth, err := userPath(name)
	if err != nil {
		return err
	}

	// Create parent directories only when the caller is allowed to create the
	// file. Update opens an existing file and must not create the users directory.
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

func readUserData(name string) ([]byte, error) {
	pth, err := userPath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return data, nil
}
