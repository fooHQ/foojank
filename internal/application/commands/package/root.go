package _package

import (
	"github.com/urfave/cli/v2"
	"log/slog"
)

func NewRootCommand(logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:        "package",
		Description: "Manage fzz packages.",
		Subcommands: []*cli.Command{
			NewBuildCommand(),
		},
		HideHelpCommand: true,
	}
}
