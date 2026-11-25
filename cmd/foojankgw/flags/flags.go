package flags

import (
	"strings"

	"github.com/urfave/cli/v3"
)

// Env is the environment source for a command flag.
// server-url becomes FOOJANKD_SERVER_URL.
func Env(name string) cli.ValueSourceChain {
	return cli.EnvVars("FOOJANK_HTTP_GATEWAY_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")))
}

const (
	ServerURL         = "server-url"
	ServerCertificate = "server-certificate"
	Credential        = "credential"
	Format            = "format"
	NoColor           = "no-color"
	Name              = "name"
	Token             = "token"
)
