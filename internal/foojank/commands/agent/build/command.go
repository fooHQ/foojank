package build

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/foohq/urlpath"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/ren/modules"
)

const (
	FlagOs               = "os"
	FlagArch             = "arch"
	FlagOutput           = "output"
	FlagDev              = "dev"
	FlagWithServer       = "with-server"
	FlagWithoutModule    = "without-module"
	FlagServer           = flags.Server
	FlagUserJWT          = flags.UserJWT
	FlagUserKey          = flags.UserKey
	FlagAccountJWT       = flags.AccountJWT
	FlagAccountKey       = flags.AccountKey
	FlagTLSCACertificate = flags.TLSCACertificate
	FlagDataDir          = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an agent",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  FlagDev,
				Usage: "enable development mode",
				Value: false,
			},
			&cli.StringFlag{
				Name:  FlagOs,
				Usage: "set build operating system",
			},
			&cli.StringFlag{
				Name:  FlagArch,
				Usage: "set build architecture",
			},
			&cli.StringFlag{
				Name:    FlagOutput,
				Usage:   "set output file",
				Aliases: []string{"o"},
			},
			&cli.StringSliceFlag{
				Name:  FlagWithServer,
				Usage: "set agent's server",
			},
			&cli.StringSliceFlag{
				Name:  FlagWithoutModule,
				Usage: "disable compilation of a module",
			},
			&cli.StringSliceFlag{
				Name:  FlagServer,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  FlagUserJWT,
				Usage: "set user JWT token",
			},
			&cli.StringFlag{
				Name:  FlagUserKey,
				Usage: "set user secret key",
			},
			&cli.StringFlag{
				Name:  FlagAccountJWT,
				Usage: "set account JWT",
			},
			&cli.StringFlag{
				Name:  FlagAccountKey,
				Usage: "set account signing key",
			},
			&cli.StringFlag{
				Name:  FlagTLSCACertificate,
				Usage: "set TLS CA certificate",
			},
			&cli.StringFlag{
				Name:  FlagDataDir,
				Usage: "set path to a data directory",
			},
		},
		Action:       action,
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey, *conf.Client.TLSCACertificate)
	if err != nil {
		err := fmt.Errorf("cannot connect to the server: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	codebaseCli := codebase.New(*conf.DataDir)
	repositoryCli := repository.New(js)
	return buildAction(logger, conf, codebaseCli, repositoryCli)(ctx, c)
}

func buildAction(logger *slog.Logger, conf *config.Config, codebaseCli *codebase.Client, repositoryCli *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		outputName := c.String(FlagOutput)
		targetOs := c.String(FlagOs)
		targetArch := c.String(FlagArch)
		devBuild := c.Bool(FlagDev)
		agentServer := c.StringSlice(FlagWithServer)
		disabledMods := c.StringSlice(FlagWithoutModule)

		agentName := nuid.Next()
		if outputName == "" {
			wd, err := os.Getwd()
			if err != nil {
				err := fmt.Errorf("cannot build an agent: cannot determine current working directory")
				logger.ErrorContext(ctx, err.Error())
				return err
			}

			outputName = filepath.Join(wd, agentName)
		}

		if targetOs == "" {
			targetOs = runtime.GOOS
		}

		if targetArch == "" {
			targetArch = runtime.GOARCH
		}

		if targetOs == "windows" && !strings.HasSuffix(outputName, ".exe") {
			outputName += ".exe"
		}

		servers := agentServer
		if len(servers) == 0 {
			err := errors.New("cannot build an agent: no server configured")
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		for i := range servers {
			scheme, err := urlpath.Scheme(servers[i])
			if err != nil {
				err := fmt.Errorf("cannot build an agent: %w", err)
				logger.ErrorContext(ctx, err.Error())
				return err
			}

			if scheme == "" {
				servers[i] = fmt.Sprintf("wss://%s", servers[i])
			}
		}

		accountClaims, err := jwt.DecodeAccountClaims(*conf.Client.AccountJWT)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot decode account JWT: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		user, err := auth.NewUserAgent(agentName, accountPubKey, []byte(*conf.Client.AccountKey))
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		agentConf := templateData{
			Servers:          servers,
			TLSCACertificate: *conf.Client.TLSCACertificate,
			User: templateDataUser{
				JWT:     user.JWT,
				KeySeed: user.Key,
			},
			Service: templateDataService{
				Name:    agentName,
				Version: foojank.Version(),
			},
		}
		confOutput, err := RenderTemplate(templateString, agentConf)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot generate agent configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = codebaseCli.WriteAgentConfig(confOutput)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot write agent configuration to a file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		mods, err := codebaseCli.ListModules()
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot get a list of modules: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		for _, mod := range disabledMods {
			if !moduleExists(mods, mod) {
				err := fmt.Errorf("cannot build an agent: module '%s' does not exist", mod)
				logger.ErrorContext(ctx, err.Error())
				return err
			}
		}

		buildTags := configureBuildTags(mods, disabledMods)
		binPath, output, err := codebaseCli.BuildAgent(ctx, targetOs, targetArch, !devBuild, buildTags)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: %w\n%s", err, output)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		repoName := agentName
		repoDescription := fmt.Sprintf("Agent %s (%s/%s)", repoName, targetOs, targetArch)
		err = repositoryCli.Create(ctx, repoName, repoDescription)
		if err != nil {
			_ = os.Remove(binPath)
			err := fmt.Errorf("cannot create repository '%s': %w", repoName, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = os.Rename(binPath, outputName)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot rename the executable file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, filepath.Base(outputName))

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	if conf.DataDir == nil {
		return errors.New("data directory not configured")
	}

	if conf.Client == nil {
		return errors.New("client configuration is missing")
	}

	if len(conf.Client.Server) == 0 {
		return errors.New("server not configured")
	}

	if conf.Client.UserJWT == nil {
		return errors.New("user jwt not configured")
	}

	if conf.Client.UserKey == nil {
		return errors.New("user key not configured")
	}

	if conf.Client.AccountJWT == nil {
		return errors.New("account jwt not configured")
	}

	if conf.Client.AccountKey == nil {
		return errors.New("account key not configured")
	}

	if conf.Client.TLSCACertificate == nil {
		return errors.New("tls ca certificate not configured")
	}

	return nil
}

func moduleExists(mods []string, name string) bool {
	for _, m := range mods {
		if m == name {
			return true
		}
	}
	return false
}

func configureBuildTags(enabledMods, disabledMods []string) []string {
	disabled := make(map[string]struct{}, len(disabledMods))
	for _, disabledMod := range disabledMods {
		disabled[disabledMod] = struct{}{}
	}

	result := make([]string, 0, len(disabled))
	for _, e := range enabledMods {
		_, isDisabled := disabled[e]
		if isDisabled {
			result = append(result, modules.StubBuildTag(e))
		}
	}

	return result
}
