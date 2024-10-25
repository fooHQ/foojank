package agent

import (
	"github.com/foojank/foojank/clients/vessel"
	"github.com/urfave/cli/v2"
	"log/slog"
)

type Arguments struct {
	Logger *slog.Logger
	Vessel *vessel.Client
}

func NewRootCommand(args Arguments) *cli.Command {
	return &cli.Command{
		Name:        "agent",
		Description: "Command & control installed agents.",
		Subcommands: []*cli.Command{
			NewListCommand(ListArguments(args)),
			NewRunCommand(RunArguments(args)),
		},
		HideHelpCommand: true,
	}
}
