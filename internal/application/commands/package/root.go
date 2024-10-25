package _package

import (
	"github.com/urfave/cli/v2"
	"log/slog"
)

type Arguments struct {
	Logger *slog.Logger
}

func NewRootCommand(args Arguments) *cli.Command {
	return &cli.Command{
		Name:        "package",
		Description: "Manage fzz packages.",
		Subcommands: []*cli.Command{
			NewBuildCommand(BuildArguments{
				Logger: args.Logger,
			}),
		},
		HideHelpCommand: true,
	}
}
