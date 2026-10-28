package authdir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"

	"github.com/foohq/foojank/internal/auth"
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

type User struct {
	*auth.User
	jwt string
}

func (u *User) JWT() string {
	return u.jwt
}

func GetUser(name string) (*User, error) {
	data, err := readUserData(name)
	if err != nil {
		return nil, err
	}

	keyPair, err := jwt.ParseDecoratedNKey(data)
	if err != nil {
		return nil, err
	}

	userJWT, err := jwt.ParseDecoratedJWT(data)
	if err != nil {
		return nil, err
	}

	claims, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return nil, err
	}

	user, err := auth.NewUser(claims, keyPair)
	if err != nil {
		return nil, err
	}

	return &User{
		jwt:  userJWT,
		User: user,
	}, nil
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
	return writeData(pth, data, flag)
}

func readUserData(name string) ([]byte, error) {
	pth, err := userPath(name)
	if err != nil {
		return nil, err
	}
	b, err := readData(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return b, nil
}
