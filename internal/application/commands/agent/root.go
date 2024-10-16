package agent

import (
	vesselcli "github.com/foojank/foojank/clients/vessel"
	"github.com/urfave/cli/v2"
)

func NewRootCommand(vessel *vesselcli.Client) *cli.Command {
	return &cli.Command{
		Name:        "agent",
		Description: "Command & control installed agents.",
		Subcommands: []*cli.Command{
			NewListCommand(vessel),
			NewRunCommand(vessel),
		},
		HideHelpCommand: true,
	}
}
