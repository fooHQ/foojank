package auth

import (
	"fmt"

	"github.com/nats-io/jwt/v2"

	protoagent "github.com/foohq/foojank/proto/agent"
	protogw "github.com/foohq/foojank/proto/gateway"
)

func NewClientPermissions(clientID string, pubPerms []string) (jwt.Permissions, error) {
	return jwt.Permissions{
		Pub: jwt.Permission{
			Allow: append([]string{
				inboxName("*") + ".>",
			}, pubPerms...),
		},
		Sub: jwt.Permission{
			Allow: []string{
				inboxName(clientID) + ".>",
			},
		},
	}, nil
}

func NewGatewayPermissions(stream, gatewayID string) jwt.Permissions {
	return jwt.Permissions{
		Pub: jwt.Permission{
			Allow: []string{
				/*"$JS.API.STREAM.INFO." + gatewayID,
				"$JS.API.STREAM.INFO.OBJ_" + gatewayID,
				"$JS.API.STREAM.PURGE.OBJ_" + gatewayID,*/
				fmt.Sprintf("$JS.API.CONSUMER.INFO.%s.%s", stream, gatewayID),
				fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.%s.%s", stream, gatewayID),
				fmt.Sprintf("$JS.ACK.%s.%s.>", stream, gatewayID),
				inboxName("*") + ".>",
				/*fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.>", gatewayID),
				fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", gatewayID),
				fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.>", gatewayID),*/
				/*fmt.Sprintf("$O.%s.M.*", gatewayID),
				fmt.Sprintf("$O.%s.C.*", gatewayID),*/
			},
		},
		Sub: jwt.Permission{
			Allow: []string{
				inboxName(gatewayID) + ".>",
				protogw.RegisterAgentSubject(gatewayID),
				protogw.UnregisterAgentSubject(gatewayID),
			},
		},
	}
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
	return "_INBOX." + name
}
