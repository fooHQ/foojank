package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const commandListName = "COMMAND.LIST"

// CommandList is the COMMAND.LIST privilege.
type CommandList struct{}

func parseCommandList(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", commandListName)
	}
	return CommandList{}, nil
}

func (CommandList) String() string {
	return commandListName
}

func (CommandList) Subject() (string, error) {
	return protodaemon.ListCommandsSubject(), nil
}
