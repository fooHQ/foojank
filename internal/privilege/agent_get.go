package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const agentGetName = "AGENT.GET"

// AgentGet is the AGENT.GET privilege.
type AgentGet struct{}

func parseAgentGet(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", agentGetName)
	}
	return AgentGet{}, nil
}

func (AgentGet) String() string {
	return agentGetName
}

func (AgentGet) Subject() (string, error) {
	return protodaemon.GetAgentSubject(), nil
}
