package auth

import (
	"github.com/nats-io/jwt/v2"

	protoagent "github.com/foohq/foojank/proto/agent"
)

func NewClientPermissions(privileges []Privilege) (jwt.Permissions, error) {
	var perms []string
	for _, priv := range privileges {
		if priv == nil {
			continue
		}
		subj, err := priv.Subject()
		if err != nil {
			return jwt.Permissions{}, err
		}
		perms = append(perms, subj)
	}
	return jwt.Permissions{
		Pub: jwt.Permission{
			Allow: perms,
		},
	}, nil
}

func NewAgentPermissions(gatewayID, agentID string) jwt.Permissions {
	return jwt.Permissions{
		Pub: jwt.Permission{
			Allow: []string{
				protoagent.EvtStartWorkerSubject(gatewayID, agentID, "*"),
				protoagent.EvtStopWorkerSubject(gatewayID, agentID, "*"),
				protoagent.EvtWorkerStdoutSubject(gatewayID, agentID, "*"),
				protoagent.EvtWorkerStatusSubject(gatewayID, agentID, "*"),
				protoagent.EvtAgentInfoSubject(gatewayID, agentID),
			},
		},
		Sub: jwt.Permission{
			Allow: []string{
				inboxName(gatewayID) + ".>",
			},
		},
	}
}

func inboxName(name string) string {
	return "_INBOX_" + name
}
