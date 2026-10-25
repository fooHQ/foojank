package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const commandGetName = "COMMAND.GET"

// CommandGet is the COMMAND.GET privilege.
type CommandGet struct{}

func parseCommandGet(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", commandGetName)
	}
	return CommandGet{}, nil
}

func (CommandGet) String() string {
	return commandGetName
}

func (CommandGet) Subject() (string, error) {
	return protodaemon.GetCommandSubject(), nil
}
