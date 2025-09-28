package build

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/foohq/ren/modules"
	"github.com/foohq/urlpath"
	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
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

	srv, err := server.New(conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey, *conf.Client.TLSCACertificate)
	if err != nil {
		log.Error(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	codebaseCli := codebase.New(*conf.DataDir)

	agentID := petname.Generate(2, "-")
	outputName := c.String(FlagOutput)
	targetOS := c.String(FlagOs)
	targetArch := c.String(FlagArch)
	isDevBuild := c.Bool(FlagDev)
	agentServer := c.StringSlice(FlagWithServer)
	disabledModules := c.StringSlice(FlagWithoutModule)

	if outputName == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Error(ctx, "Cannot get current working directory")
			return err
		}

		outputName = filepath.Join(wd, agentID)
		if targetOS == "windows" && !strings.HasSuffix(outputName, ".exe") {
			outputName += ".exe"
		}
	}

	if targetOS == "" {
		targetOS = runtime.GOOS
	}

	if targetArch == "" {
		targetArch = runtime.GOARCH
	}

	buildMode := "production"
	if isDevBuild {
		buildMode = "development"
	}

	servers := agentServer
	if len(servers) == 0 {
		log.Error(ctx, "No server configured. Use --%s flag to specify one", FlagWithServer)
		return err
	}

	for i := range servers {
		scheme, err := urlpath.Scheme(servers[i])
		if err != nil {
			return err
		}

		if scheme == "" {
			servers[i] = fmt.Sprintf("wss://%s", servers[i])
		}
	}

	streamName := vessel.StreamName(agentID)

	log.Debug(ctx, "Create a stream")

	_, err = srv.CreateStream(ctx, streamName, []string{
		fmt.Sprintf(vessel.SubjectApiWorkerStartT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiWorkerStopT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiWorkerWriteStdinT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiWorkerWriteStdoutT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiWorkerStatusT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiReplyT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiConnInfoT, agentID),
	})
	if err != nil {
		log.Error(ctx, "Cannot create stream: %v", err)
		return err
	}

	consumerName := agentID

	log.Debug(ctx, "Create a message consumer")

	_, err = srv.CreateDurableConsumer(ctx, streamName, consumerName, []string{
		fmt.Sprintf(vessel.SubjectApiWorkerStartT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiWorkerStopT, agentID, "*"),
		fmt.Sprintf(vessel.SubjectApiWorkerWriteStdinT, agentID, "*"),
	})
	if err != nil {
		log.Error(ctx, "Cannot create consumer: %v", err)
		return err
	}

	storeName := agentID
	storeDescription := fmt.Sprintf("Agent %s (%s/%s)", agentID, targetOS, targetArch)

	log.Info(ctx, "Create a store")

	err = srv.CreateObjectStore(ctx, storeName, storeDescription)
	if err != nil {
		log.Error(ctx, "Cannot create a store: %v", err)
		return err
	}

	usr, err := createUser(*conf.Client.AccountJWT, *conf.Client.AccountKey, agentID)
	if err != nil {
		log.Error(ctx, "Cannot create a user: %v", err)
		return err
	}

	log.Info(ctx, "Build an executable file %q [%s/%s] [%s]", outputName, targetOS, targetArch, buildMode)

	err = buildExecutable(
		ctx,
		codebaseCli,
		targetOS,
		targetArch,
		outputName,
		isDevBuild,
		disabledModules,
		codebase.BuildAgentConfig{
			ID:                           agentID,
			Server:                       strings.Join(servers, ","),
			UserJWT:                      usr.JWT,
			UserKey:                      usr.Key,
			CACertificate:                *conf.Client.TLSCACertificate,
			Stream:                       streamName,
			Consumer:                     consumerName,
			InboxPrefix:                  vessel.InboxName(agentID),
			ObjectStoreName:              agentID,
			SubjectApiWorkerStartT:       vessel.SubjectApiWorkerStartT,
			SubjectApiWorkerStopT:        vessel.SubjectApiWorkerStopT,
			SubjectApiWorkerWriteStdinT:  vessel.SubjectApiWorkerWriteStdinT,
			SubjectApiWorkerWriteStdoutT: vessel.SubjectApiWorkerWriteStdoutT,
			SubjectApiWorkerStatusT:      vessel.SubjectApiWorkerStatusT,
			SubjectApiConnInfoT:          vessel.SubjectApiConnInfoT,
			SubjectApiReplyT:             vessel.SubjectApiReplyT,
			ReconnectInterval:            1 * time.Minute, // TODO: make configurable
			ReconnectJitter:              5 * time.Second, // TODO: make configurable
			AwaitMessagesDuration:        5 * time.Second, // TODO: make configurable
		},
	)
	if err != nil {
		log.Error(ctx, "Cannot build executable: %v", err)
		return err
	}

	log.Info(ctx, "Agent %q has been built!", agentID)

	return nil
}

func createUser(
	accountJWT,
	accountKey,
	agentID string,
) (*auth.User, error) {
	accountClaims, err := jwt.DecodeAccountClaims(accountJWT)
	if err != nil {
		return nil, fmt.Errorf("cannot decode account JWT: %w", err)
	}

	accountPubKey := accountClaims.Subject
	user, err := auth.NewUserAgent(agentID, accountPubKey, []byte(accountKey))
	if err != nil {
		return nil, fmt.Errorf("cannot generate agent configuration: %w", err)
	}

	return user, nil
}

func buildExecutable(
	ctx context.Context,
	cli *codebase.Client,
	targetOS,
	targetArch,
	outputName string,
	isDevBuild bool,
	disabledModules []string,
	conf codebase.BuildAgentConfig,
) error {
	buildTags, err := configureBuildTags(modules.Modules(), disabledModules)
	if err != nil {
		return fmt.Errorf("cannot configure modules: %w", err)
	}

	binPath, output, err := cli.BuildAgent(ctx, codebase.BuildAgentOptions{
		OS:         targetOS,
		Arch:       targetArch,
		Production: !isDevBuild,
		Tags:       buildTags,
		Config:     conf,
	})
	if err != nil {
		return fmt.Errorf("%s", output)
	}

	err = os.Rename(binPath, outputName)
	if err != nil {
		_ = os.Remove(binPath)
		return fmt.Errorf("cannot rename the executable file: %w", err)
	}

	// TODO: should be debug!
	log.Info(ctx, "Go flags: %s", conf.ToFlags())

	return nil
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

func configureBuildTags(mods, disabledMods []string) ([]string, error) {
	// Verify that disabled modules exist, otherwise throw an error.
	for _, m := range disabledMods {
		if !moduleExists(mods, m) {
			err := fmt.Errorf("module '%s' does not exist", m)
			return nil, err
		}
	}

	var buildTags []string
	for _, m := range mods {
		if moduleExists(disabledMods, m) {
			buildTags = append(buildTags, modules.StubBuildTag(m))
		}
	}
	return buildTags, nil
}
