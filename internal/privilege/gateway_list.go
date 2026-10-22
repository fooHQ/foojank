package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const gatewayListName = "GATEWAY.LIST"

// GatewayList is the GATEWAY.LIST privilege.
type GatewayList struct{}

func parseGatewayList(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", gatewayListName)
	}
	return GatewayList{}, nil
}

func (GatewayList) String() string {
	return gatewayListName
}

func (GatewayList) Subject() (string, error) {
	return protodaemon.ListGatewaysSubject(), nil
}
