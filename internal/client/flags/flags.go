package flags

import (
	"os"
	"path/filepath"
)

const (
	Config            = "config"
	Server            = "server"
	UserJWT           = "user-jwt"
	UserKey           = "user-key"
	AccountJWT        = "account-jwt"
	AccountSigningKey = "account-signing-key"
	LogLevel          = "log-level"
	NoColor           = "no-color"
	Codebase          = "codebase"
)

var (
	DefaultConfig = func() string {
		confDir, err := os.UserConfigDir()
		if err != nil {
			confDir = "./"
		}
		return filepath.Join(confDir, "foojank", "client.conf")
	}
	DefaultServer         = []string{"localhost:4222"}
	DefaultLogLevel int64 = 0
	DefaultNoColor        = false
)
