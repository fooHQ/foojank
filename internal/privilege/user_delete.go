package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const userDeleteName = "USER.DELETE"

// UserDelete is the USER.DELETE privilege.
type UserDelete struct{}

func parseUserDelete(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", userDeleteName)
	}
	return UserDelete{}, nil
}

func (UserDelete) String() string {
	return userDeleteName
}

func (UserDelete) Subject() (string, error) {
	return protodaemon.DeleteUserSubject(), nil
}
