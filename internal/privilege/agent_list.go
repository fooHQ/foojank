package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const agentListName = "AGENT.LIST"

// AgentList is the AGENT.LIST privilege.
type AgentList struct{}

func parseAgentList(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", agentListName)
	}
	return AgentList{}, nil
}

func (AgentList) String() string {
	return agentListName
}

func (AgentList) Subject() (string, error) {
	return protodaemon.ListAgentsSubject(), nil
}
