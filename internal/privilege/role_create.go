package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const roleCreateName = "ROLE.CREATE"

// RoleCreate is the ROLE.CREATE privilege.
type RoleCreate struct{}

func parseRoleCreate(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", roleCreateName)
	}
	return RoleCreate{}, nil
}

func (RoleCreate) String() string {
	return roleCreateName
}

func (RoleCreate) Subject() (string, error) {
	return protodaemon.CreateRoleSubject(), nil
}
