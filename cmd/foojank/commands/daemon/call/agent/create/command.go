package create

import (
	"context"
	"errors"
	"maps"
	"os"
	"runtime"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create an agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set agent name",
			},
			&cli.StringFlag{
				Name:  flags.Description,
				Usage: "set agent description",
			},
			&cli.StringFlag{
				Name:  flags.Gateway,
				Usage: "set agent's gateway",
			},
			&cli.StringFlag{
				Name:  flags.Profile,
				Usage: "set a profile to use",
			},
			&cli.StringFlag{
				Name:  flags.Os,
				Usage: "set target operating system",
			},
			&cli.StringFlag{
				Name:  flags.Arch,
				Usage: "set target architecture",
			},
			&cli.StringSliceFlag{
				Name:  flags.Variable,
				Usage: "set config variable (format: key=value)",
			},
			&cli.StringSliceFlag{
				Name:  flags.WithoutVariable,
				Usage: "unset config variable (format: key)",
			},
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set server account",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
		ShellComplete:   actions.CompleteFlags,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.LoadProfiles(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, _ *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	profs := actions.GetProfilesFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	targetOS, _ := conf.String(flags.Os)
	targetArch, _ := conf.String(flags.Arch)
	setVars, _ := conf.StringSlice(flags.Variable)
	unsetVars, _ := conf.StringSlice(flags.WithoutVariable)
	gatewayName, _ := conf.String(flags.Gateway)
	agentName, _ := conf.String(flags.Name)
	agentDesc, _ := conf.String(flags.Description)
	profName, _ := conf.String(flags.Profile)

	userJWT, userSeed, err := authdir.ReadUser(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", accountName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	client := daemon.New(srv)

	if agentName == "" {
		agentName = petname.Generate(2, "-")
	}

	agentConf := protodaemon.AgentConfig{
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
		Extra: make(map[string]string),
	}

	// Parse profile file.
	if profName != "" {
		prof, err := profs.Get(profName)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get profile %q: %v", profName, err)
			return err
		}

		v := prof.OS()
		if v != "" {
			agentConf.OS = v
		}

		v = prof.Arch()
		if v != "" {
			agentConf.Arch = v
		}

		maps.Copy(agentConf.Extra, prof.Env())
	}

	// Parse command line flags.
	if targetOS != "" {
		agentConf.OS = targetOS
	}

	if targetArch != "" {
		agentConf.Arch = targetArch
	}

	maps.Copy(agentConf.Extra, parseKVPairs(setVars))

	for _, v := range unsetVars {
		delete(agentConf.Extra, v)
	}

	_, err = client.RequestCreateAgent(ctx, protodaemon.CreateAgentRequest{
		Name:        agentName,
		Description: agentDesc,
		Gateway:     gatewayName,
		Config:      agentConf,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create agent: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Agent %q has been created!", agentName)

	return nil
}

func parseKVPairs(pairs []string) map[string]string {
	env := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		var v string
		if len(parts) > 1 {
			v = parts[1]
		}
		env[strings.TrimSpace(parts[0])] = v
	}
	return env
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Gateway,
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
		case flags.Gateway:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("gateway not configured")
			}
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
