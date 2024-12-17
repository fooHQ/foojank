package flags

import (
	"os"
	"path/filepath"
)

const (
	Config           = "config"
	Host             = "host"
	Port             = "port"
	OperatorJWT      = "operator-jwt"
	AccountJWT       = "account-jwt"
	SystemAccountJWT = "sys-account-jwt"
	LogLevel         = "log-level"
	NoColor          = "no-color"
)

var (
	DefaultConfig = func() string {
		confDir, err := os.UserConfigDir()
		if err != nil {
			confDir = "./"
		}
		return filepath.Join(confDir, "foojank", "server.conf")
	}
	DefaultHost           = "localhost"
	DefaultPort     int64 = 4222
	DefaultLogLevel int64 = 0
	DefaultNoColor        = false
)
