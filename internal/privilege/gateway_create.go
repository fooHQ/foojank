package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const gatewayCreateName = "GATEWAY.CREATE"

// GatewayCreate is the GATEWAY.CREATE privilege.
type GatewayCreate struct{}

func parseGatewayCreate(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", gatewayCreateName)
	}
	return GatewayCreate{}, nil
}

func (GatewayCreate) String() string {
	return gatewayCreateName
}

func (GatewayCreate) Subject() (string, error) {
	return protodaemon.CreateGatewaySubject(), nil
}
