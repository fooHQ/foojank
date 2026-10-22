package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const gatewayDeleteName = "GATEWAY.DELETE"

// GatewayDelete is the GATEWAY.DELETE privilege.
type GatewayDelete struct{}

func parseGatewayDelete(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", gatewayDeleteName)
	}
	return GatewayDelete{}, nil
}

func (GatewayDelete) String() string {
	return gatewayDeleteName
}

func (GatewayDelete) Subject() (string, error) {
	return protodaemon.DeleteGatewaySubject(), nil
}
