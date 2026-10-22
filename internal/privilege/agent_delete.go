package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const agentDeleteName = "AGENT.DELETE"

// AgentDelete is the AGENT.DELETE privilege.
type AgentDelete struct{}

func parseAgentDelete(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", agentDeleteName)
	}
	return AgentDelete{}, nil
}

func (AgentDelete) String() string {
	return agentDeleteName
}

func (AgentDelete) Subject() (string, error) {
	return protodaemon.DeleteAgentSubject(), nil
}
