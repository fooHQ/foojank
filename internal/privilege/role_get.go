package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const roleGetName = "ROLE.GET"

// RoleGet is the ROLE.GET privilege.
type RoleGet struct{}

func parseRoleGet(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", roleGetName)
	}
	return RoleGet{}, nil
}

func (RoleGet) String() string {
	return roleGetName
}

func (RoleGet) Subject() (string, error) {
	return protodaemon.GetRoleSubject(), nil
}
