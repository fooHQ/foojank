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
	ErrOperatorNotFound = errors.New("operator not found")
	ErrOperatorExists   = errors.New("operator already exists")
)

func CreateOperator(name string, operatorJWT string, operatorSeed []byte) error {
	if !isOperatorJWT(operatorJWT) {
		return errors.New("invalid operator JWT")
	}

	if !isOperatorSeed(operatorSeed) {
		return errors.New("invalid operator seed")
	}

	data, err := decorate(operatorJWT, operatorSeed)
	if err != nil {
		return err
	}

	err = writeOperatorData(name, data, os.O_CREATE|os.O_WRONLY|os.O_EXCL)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrOperatorExists
		}
		return err
	}

	return nil
}

func UpdateOperator(name string, operatorJWT string, operatorSeed []byte) error {
	if !isOperatorJWT(operatorJWT) {
		return errors.New("invalid operator JWT")
	}

	if !isOperatorSeed(operatorSeed) {
		return errors.New("invalid operator seed")
	}

	data, err := decorate(operatorJWT, operatorSeed)
	if err != nil {
		return err
	}

	err = writeOperatorData(name, data, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrOperatorNotFound
		}
		return err
	}

	return nil
}

func GetOperatorKey(name string) (nkeys.KeyPair, error) {
	data, err := readOperatorData(name)
	if err != nil {
		return nil, err
	}
	return jwt.ParseDecoratedNKey(data)
}

func GetOperatorJWT(name string) (*jwt.OperatorClaims, error) {
	data, err := readOperatorData(name)
	if err != nil {
		return nil, err
	}

	operatorJWT, err := jwt.ParseDecoratedJWT(data)
	if err != nil {
		return nil, err
	}

	return jwt.DecodeOperatorClaims(operatorJWT)
}

func ReadOperator(name string) (string, []byte, error) {
	data, err := readOperatorData(name)
	if err != nil {
		return "", nil, err
	}

	operatorJWT, err := jwt.ParseDecoratedJWT(data)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode decorated JWT: %w", err)
	}

	// Validate that the JWT is an operator JWT.
	_, err = jwt.DecodeOperatorClaims(operatorJWT)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode JWT: %w", err)
	}

	operator, err := jwt.ParseDecoratedNKey(data)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode decorated seed: %w", err)
	}

	operatorSeed, err := operator.Seed()
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode seed: %w", err)
	}

	if !isOperatorSeed(operatorSeed) {
		return "", nil, errors.New("invalid operator seed")
	}

	return operatorJWT, operatorSeed, nil
}

func ListOperators() ([]string, error) {
	pth, err := operatorRootPath()
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

	var operators []string
	for _, file := range files {
		operators = append(operators, file.Name())
	}

	return operators, nil
}

func GetOperatorPath(name string) (string, error) {
	pth, err := operatorPath(name)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrOperatorNotFound
		}
		return "", err
	}

	return pth, nil
}

func DeleteOperator(name string) error {
	pth, err := operatorPath(name)
	if err != nil {
		return err
	}

	err = os.RemoveAll(pth)
	if err != nil {
		return err
	}
	return nil
}

func operatorRootPath() (string, error) {
	root, err := rootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "operators"), nil
}

func operatorPath(name string) (string, error) {
	name = strings.ToLower(name)
	root, err := operatorRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, name), nil
}

func isOperatorJWT(s string) bool {
	_, err := jwt.DecodeOperatorClaims(s)
	return err == nil
}

func isOperatorSeed(key []byte) bool {
	kp, err := nkeys.FromSeed(key)
	if err != nil {
		return false
	}

	pub, err := kp.PublicKey()
	if err != nil {
		return false
	}

	return nkeys.IsValidPublicOperatorKey(pub)
}

func writeOperatorData(name string, data []byte, flag int) error {
	pth, err := operatorPath(name)
	if err != nil {
		return err
	}
	return writeData(pth, data, flag)
}

func readOperatorData(name string) ([]byte, error) {
	pth, err := operatorPath(name)
	if err != nil {
		return nil, err
	}
	b, err := readData(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrOperatorNotFound
		}
		return nil, err
	}
	return b, nil
}
