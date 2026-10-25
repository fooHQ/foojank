package actions

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func TestCompleteJobUsage(t *testing.T) {
	defined := map[string]protodaemon.AgentCommand{
		"ls": {
			Description: "List a directory",
			Flags: map[string]protodaemon.AgentCommandFlag{
				"all":       {Description: "Include hidden entries"},
				"recursive": {Description: "Recurse into directories"},
				"v":         {Description: "Verbose"},
				"agent":     {Description: "Select an agent"},
			},
		},
		"ps": {
			Description: "List processes",
		},
	}
	list := func(context.Context) (map[string]protodaemon.AgentCommand, error) {
		return defined, nil
	}
	listErr := func(context.Context) (map[string]protodaemon.AgentCommand, error) {
		return nil, errors.New("missing")
	}

	tests := []struct {
		name     string
		args     []string
		commands func(context.Context) (map[string]protodaemon.AgentCommand, error)
		want     string
	}{
		{
			name: "command names before a command is specified",
			args: []string{"create", "--agent", "bot", "--generate-shell-completion"},
			want: "ls:List a directory\nps:List processes\n",
		},
		{
			name: "partial foojank flag before a command is specified",
			args: []string{"create", "--var", "--generate-shell-completion"},
			want: "--variable:set environment variable (format: key=value)\n",
		},
		{
			name: "cli flag after the agent value and before the command",
			args: []string{"create", "--agent", "bot", "--var", "--generate-shell-completion"},
			want: "--variable:set environment variable (format: key=value)\n",
		},
		{
			name: "command flags after the command name",
			args: []string{"create", "--agent", "bot", "ls", "--generate-shell-completion"},
			want: "--agent:Select an agent\n--all:Include hidden entries\n--recursive:Recurse into directories\n-v:Verbose\n",
		},
		{
			name: "command flag prefix",
			args: []string{"create", "--agent", "bot", "ls", "--re", "--generate-shell-completion"},
			want: "--recursive:Recurse into directories\n",
		},
		{
			name: "command flag prefix excludes cli flags",
			args: []string{"create", "--agent", "bot", "ls", "--a", "--generate-shell-completion"},
			want: "--agent:Select an agent\n--all:Include hidden entries\n",
		},
		{
			name: "--variable before the command suggests no values",
			args: []string{"create", "--agent", "bot", "--variable", "--generate-shell-completion"},
			want: "",
		},
		{
			name: "--variable= before the command suggests no values",
			args: []string{"create", "--agent", "bot", "--variable=D", "--generate-shell-completion"},
			want: "",
		},
		{
			name: "--variable after the command suggests no values",
			args: []string{"create", "--agent", "bot", "ls", "--variable", "--generate-shell-completion"},
			want: "",
		},
		{
			name:     "missing command definition prints no flags",
			args:     []string{"create", "--agent", "bot", "ls", "--generate-shell-completion"},
			commands: listErr,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := os.Args
			t.Cleanup(func() { os.Args = orig })
			os.Args = tt.args

			commands := tt.commands
			if commands == nil {
				commands = list
			}

			var out bytes.Buffer
			cmd := &cli.Command{
				Name:                  "create",
				EnableShellCompletion: true,
				HideHelp:              true,
				HideVersion:           true,
				StopOnNthArg:          new(1),
				Writer:                &out,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  flags.Agent,
						Usage: "assign the job to the specified agent",
					},
					&cli.StringSliceFlag{
						Name:  flags.Variable,
						Usage: "set environment variable (format: key=value)",
					},
				},
				ShellComplete: func(ctx context.Context, c *cli.Command) {
					completeJobUsage(ctx, c, commands)
				},
				Action: func(context.Context, *cli.Command) error { return nil },
			}

			err := cmd.Run(context.Background(), tt.args)
			require.NoError(t, err)
			require.Equal(t, tt.want, out.String())
		})
	}
}
