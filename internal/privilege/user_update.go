package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const userUpdateName = "USER.UPDATE"

// UserUpdate is the USER.UPDATE privilege.
type UserUpdate struct{}

func parseUserUpdate(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", userUpdateName)
	}
	return UserUpdate{}, nil
}

func (UserUpdate) String() string {
	return userUpdateName
}

func (UserUpdate) Subject() (string, error) {
	return protodaemon.UpdateUserSubject(), nil
}
