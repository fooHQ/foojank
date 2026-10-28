package authdir

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nats-io/jwt/v2"
)

// Root is the directory that contains the accounts, users, and operators stores.
// An empty value uses the default, UserConfigDir()/foojank.
// Set Root before calling any other function in this package.
var Root string

func rootPath() (string, error) {
	if Root != "" {
		return Root, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "foojank"), nil
}

func decorate(sJWT string, sSeed []byte) ([]byte, error) {
	jwtDecorated, err := jwt.DecorateJWT(sJWT)
	if err != nil {
		return nil, fmt.Errorf("cannot decorate JWT: %w", err)
	}

	seedDecorated, err := jwt.DecorateSeed(sSeed)
	if err != nil {
		return nil, fmt.Errorf("cannot decorate seed: %w", err)
	}

	return bytes.Join([][]byte{jwtDecorated, seedDecorated}, []byte("")), nil
}

func writeData(pth string, data []byte, flag int) error {
	// Create parent directories only when the caller is allowed to create the
	// file. Update opens an existing file; creating the parent first would
	// leave an empty directory that ListAccounts treats as an account.
	if flag&os.O_CREATE != 0 {
		err := os.MkdirAll(filepath.Dir(pth), 0o700)
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

func readData(pth string) ([]byte, error) {
	data, err := os.ReadFile(pth)
	if err != nil {
		return nil, err
	}
	return data, nil
}
