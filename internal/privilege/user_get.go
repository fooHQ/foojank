package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const userGetName = "USER.GET"

// UserGet is the USER.GET privilege.
type UserGet struct{}

func parseUserGet(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", userGetName)
	}
	return UserGet{}, nil
}

func (UserGet) String() string {
	return userGetName
}

func (UserGet) Subject() (string, error) {
	return protodaemon.GetUserSubject(), nil
}
