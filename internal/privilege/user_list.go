package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const userListName = "USER.LIST"

// UserList is the USER.LIST privilege.
type UserList struct{}

func parseUserList(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", userListName)
	}
	return UserList{}, nil
}

func (UserList) String() string {
	return userListName
}

func (UserList) Subject() (string, error) {
	return protodaemon.ListUsersSubject(), nil
}
