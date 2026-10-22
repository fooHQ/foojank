package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const roleListName = "ROLE.LIST"

// RoleList is the ROLE.LIST privilege.
type RoleList struct{}

func parseRoleList(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", roleListName)
	}
	return RoleList{}, nil
}

func (RoleList) String() string {
	return roleListName
}

func (RoleList) Subject() (string, error) {
	return protodaemon.ListRolesSubject(), nil
}
