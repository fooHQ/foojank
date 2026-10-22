package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const gatewayGetName = "GATEWAY.GET"

// GatewayGet is the GATEWAY.GET privilege.
type GatewayGet struct{}

func parseGatewayGet(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", gatewayGetName)
	}
	return GatewayGet{}, nil
}

func (GatewayGet) String() string {
	return gatewayGetName
}

func (GatewayGet) Subject() (string, error) {
	return protodaemon.GetGatewaySubject(), nil
}
