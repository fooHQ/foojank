package flags

import (
	"strings"

	"github.com/urfave/cli/v3"
)

// Env is the environment source for a command flag.
// server-url becomes FOOJANKD_SERVER_URL.
func Env(name string) cli.ValueSourceChain {
	return cli.EnvVars("FOOJANKD_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")))
}

const (
	ServerURL         = "server-url"
	ServerCertificate = "server-certificate"
	Account           = "account"
	Format            = "format"
	Name              = "name"
	Description       = "description"
	NoColor           = "no-color"
	PublicKey         = "public-key"
	Privilege         = "privilege"
	WithoutPrivilege  = "without-privilege"
	Role              = "role"
	File              = "file"
)
