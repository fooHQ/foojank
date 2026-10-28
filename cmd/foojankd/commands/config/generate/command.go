package generate

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate server configuration",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.SystemAccount,
				Usage:   "set system account",
				Sources: flags.Env(flags.SystemAccount),
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

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupOperator(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	operator := actions.GetOperatorFromContext(ctx)

	sysAccountName, _ := conf.String(flags.SystemAccount)

	account, err := authdir.GetAccount(sysAccountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get account: %v", err)
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s", formatConfig(operator.JWT(), account.JWT(), account.PublicKey()))

	return nil
}

func formatConfig(operatorJWT, accountJWT, accountID string) string {
	return fmt.Sprintf(`operator: %s
system_account: %s

resolver {
	type: full
	# Directory in which the account jwt will be stored
	dir: './jwt'
	# In order to support jwt deletion, set to true
	# If the resolver type is full delete will rename the jwt.
	# This is to allow manual restoration in case of inadvertent deletion.
	# To restore a jwt, remove the added suffix .delete and restart or send a reload signal.
	# To free up storage you must manually delete files with the suffix .delete.
	allow_delete: false
	# Interval at which a nats-server with a nats based account resolver will compare
	# it's state with one random nats based account resolver in the cluster and if needed,
	# exchange jwt and converge on the same set of jwt.
	interval: "2m"
	# Timeout for lookup requests in case an account does not exist locally.
	timeout: "1.9s"
}

# Preload the nats based resolver with the system account jwt.
# This is not necessary but avoids a bootstrapping system account.
# This only applies to the system account. Therefore other account jwt are not included here.
# To populate the resolver:
# 1) make sure that your operator has the account server URL pointing at your nats servers.
#    The url must start with: "nats://"
#    nsc edit operator --account-jwt-server-url nats://localhost:4222
# 2) push your accounts using: nsc push --all
#    The argument to push -u is optional if your account server url is set as described.
# 3) to prune accounts use: nsc push --prune
#    In order to enable prune you must set above allow_delete to true
# Later changes to the system account take precedence over the system account jwt listed here.
resolver_preload: {
	%s: %s
}`, operatorJWT, accountID, accountID, accountJWT)
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.SystemAccount,
	} {
		switch opt {
		case flags.SystemAccount:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("system account not configured")
			}
		}
	}
	return nil
}
