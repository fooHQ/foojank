package build

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/builder"
	"github.com/foohq/foojank/internal/clients/agent"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/profile"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Profile,
				Usage: "set profile",
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
				Name:  flags.Feature,
				Usage: "enable build features",
			},
			&cli.StringFlag{
				Name:  flags.Server,
				Usage: "set agent's server URL",
			},
			&cli.StringFlag{
				Name:      flags.Certificate,
				Usage:     "set path to agent's server certificate",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:    flags.Output,
				Usage:   "set path to an output file",
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name:  flags.SourceDir,
				Usage: "set path to a source code directory",
			},
			&cli.StringSliceFlag{
				Name:  flags.Set,
				Usage: "set environment variable (format: key=value)",
			},
			&cli.StringSliceFlag{
				Name:  flags.Unset,
				Usage: "unset environment variable (format: key)",
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

func action(ctx context.Context, c *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	profs := actions.GetProfilesFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	outputName, _ := conf.String(flags.Output)
	targetOS, _ := conf.String(flags.Os)
	targetArch, _ := conf.String(flags.Arch)
	sourceDir, _ := conf.String(flags.SourceDir)
	setVars, _ := conf.StringSlice(flags.Set)
	unsetVars, _ := conf.StringSlice(flags.Unset)
	features, _ := conf.StringSlice(flags.Feature)
	targetServer, _ := conf.String(flags.Server)
	targetCert, _ := conf.String(flags.Certificate)
	profName, _ := conf.String(flags.Profile)

	_, accountSeed, err := auth.ReadAccount(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read account %q: %v", accountName, err)
		return err
	}

	userJWT, userSeed, err := auth.ReadUser(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", accountName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	client := agent.New(srv)

	agentID := petname.Generate(2, "-")

	agentJWT, agentSeed, err := createUser(agentID, string(accountSeed))
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create a user: %v", err)
		return err
	}

	pwd, err := filepath.Abs(".")
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get current directory: %v", err)
		return err
	}

	// Default profile.
	profDefault := profile.New()
	profDefault.Set(profile.VarAgentID, profile.NewVar(agentID))
	profDefault.Set(profile.VarServerURL, profile.NewVar(serverURL))
	profDefault.Set(profile.VarServerCertificate, profile.NewVar(serverCert))
	profDefault.Set(profile.VarUserJWT, profile.NewVar(agentJWT))
	profDefault.Set(profile.VarUserKey, profile.NewVar(agentSeed))
	profDefault.Set(profile.VarStream, profile.NewVar(agent.StreamName(agentID)))
	profDefault.Set(profile.VarConsumer, profile.NewVar(agentID))
	profDefault.Set(profile.VarInboxPrefix, profile.NewVar(agent.InboxName(agentID)))
	profDefault.Set(profile.VarObjectStore, profile.NewVar(agentID))
	profDefault.Set(profile.VarOS, profile.NewVar(runtime.GOOS))
	profDefault.Set(profile.VarArch, profile.NewVar(runtime.GOARCH))
	profDefault.Set(profile.VarTarget, profile.NewVar(filepath.Join(pwd, agentID)))
	profDefault.Set(profile.VarFeatures, profile.NewVar(""))

	// Parse profile profName defined in the config dir.
	var profFile *profile.Profile
	if profName != "" {
		profFile, err = profs.Get(profName)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get profile %q: %v", profName, err)
			return err
		}
	}

	// Parse profile flags.
	profFlags := profile.New()
	if sourceDir != "" {
		profFlags.SetSourceDir(sourceDir)
	}
	if targetServer != "" {
		profFlags.Set(profile.VarServerURL, profile.NewVar(targetServer))
	}
	if targetCert != "" {
		profFlags.Set(profile.VarServerCertificate, profile.NewVar(targetCert))
	}
	if targetOS != "" {
		profFlags.Set(profile.VarOS, profile.NewVar(targetOS))
	}
	if targetArch != "" {
		profFlags.Set(profile.VarArch, profile.NewVar(targetArch))
	}
	if outputName != "" {
		profFlags.Set(profile.VarTarget, profile.NewVar(filepath.Join(pwd, outputName)))
	}
	if len(features) > 0 {
		profFlags.Set(profile.VarFeatures, profile.NewVar(strings.Join(features, ",")))
	}

	for k, v := range profile.ParseKVPairs(setVars) {
		profFlags.Set(k, v)
	}

	// Merge all profiles to create a final profile.
	prof := profile.Merge(profDefault, profFile, profFlags)

	// Remove unset variables.
	for _, v := range unsetVars {
		prof.Delete(v)
	}

	sourceDir = prof.SourceDir()
	outputName = prof.Get(profile.VarTarget).Value()
	agentID = prof.Get(profile.VarAgentID).Value()

	output, err := builder.Run(ctx, sourceDir, prof.List())
	if err != nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			logger.ErrorContext(ctx, "%s", line)
		}
		logger.ErrorContext(ctx, "Build failed: %v", err)
		// Return a generic error instead of "err".
		// Err can be of type exit.ExitError, which is apparently printed to stderr by cli.
		return errors.New("build failed")
	}
	defer func() {
		if err == nil {
			return
		}
		err := os.Remove(outputName)
		if err != nil {
			logger.WarnContext(ctx, "Cannot remove executable file %q: %v", outputName, err)
		}
	}()

	err = client.Register(ctx, agentID)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot register agent: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Agent %q has been built!", agentID)

	return nil
}

func createUser(
	agentID,
	accountSeed string,
) (string, string, error) {
	perms := agent.NewAgentPermissions(agentID)
	userJWT, userSeed, err := auth.NewUser(agentID, []byte(accountSeed), perms)
	if err != nil {
		return "", "", err
	}
	return userJWT, string(userSeed), nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
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
