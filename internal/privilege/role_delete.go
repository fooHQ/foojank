package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const roleDeleteName = "ROLE.DELETE"

// RoleDelete is the ROLE.DELETE privilege.
type RoleDelete struct{}

func parseRoleDelete(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", roleDeleteName)
	}
	return RoleDelete{}, nil
}

func (RoleDelete) String() string {
	return roleDeleteName
}

func (RoleDelete) Subject() (string, error) {
	return protodaemon.DeleteRoleSubject(), nil
}
