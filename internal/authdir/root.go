package authdir

import (
	"os"
	"path/filepath"
)

// Root is the directory that contains the accounts and users stores.
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
