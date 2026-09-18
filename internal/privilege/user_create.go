package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const userCreateName = "USER.CREATE"

// UserCreate is the USER.CREATE privilege.
type UserCreate struct{}

func parseUserCreate(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", userCreateName)
	}
	return UserCreate{}, nil
}

func (UserCreate) String() string {
	return userCreateName
}

func (UserCreate) Subject() (string, error) {
	return protodaemon.CreateUserSubject(), nil
}
