package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const agentCreateName = "AGENT.CREATE"

// AgentCreate is the AGENT.CREATE privilege.
type AgentCreate struct{}

func parseAgentCreate(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", agentCreateName)
	}
	return AgentCreate{}, nil
}

func (AgentCreate) String() string {
	return agentCreateName
}

func (AgentCreate) Subject() (string, error) {
	return protodaemon.CreateAgentSubject(), nil
}
