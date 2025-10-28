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

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an agent",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  flags.Dev,
				Usage: "enable development mode",
			},
			&cli.StringFlag{
				Name:  flags.Os,
				Usage: "set build operating system",
			},
			&cli.StringFlag{
				Name:  flags.Arch,
				Usage: "set build architecture",
			},
			&cli.StringFlag{
				Name:  flags.SourceDirectory,
				Usage: "set path to source directory",
			},
			&cli.StringFlag{
				Name:    flags.Output,
				Usage:   "set output file",
				Aliases: []string{"o"},
			},
			&cli.StringSliceFlag{
				Name:  flags.WithServer,
				Usage: "set agent's server",
			},
			&cli.StringSliceFlag{
				Name:  flags.WithoutModule,
				Usage: "disable compilation of a module",
			},
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  flags.ServerCertificate,
				Usage: "set server TLS certificate",
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set server account",
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

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	outputName, _ := conf.String(flags.Output)
	targetOS, _ := conf.String(flags.Os)
	targetArch, _ := conf.String(flags.Arch)
	isDevBuild, _ := conf.Bool(flags.Dev)
	agentServer, _ := conf.StringSlice(flags.WithServer)
	disabledModules, _ := conf.StringSlice(flags.WithoutModule)
	sourceDir, _ := conf.String(flags.SourceDirectory)

	_, accountSeed, err := auth.ReadUser(accountName)
	if err != nil {
		log.Error(ctx, "Cannot read account %q: %v", accountName, err)
		return err
	}

	userJWT, userSeed, err := auth.ReadUser(accountName)
	if err != nil {
		log.Error(ctx, "Cannot read user %q: %v", accountName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		log.Error(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	codebaseCli := codebase.New(sourceDir)

	agentID := petname.Generate(2, "-")

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
		log.Error(ctx, "No server configured. Use --%s flag to specify one", flags.WithServer)
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

	agentJWT, agentSeed, err := createUser(agentID, string(accountSeed))
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
			UserJWT:                      agentJWT,
			UserKey:                      agentSeed,
			CACertificate:                serverCert,
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
	agentID,
	accountSeed string,
) (string, string, error) {
	perms := vessel.NewAgentPermissions(agentID)
	userJWT, userSeed, err := auth.NewUser(agentID, []byte(accountSeed), perms)
	if err != nil {
		return "", "", err
	}
	return userJWT, string(userSeed), nil
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
