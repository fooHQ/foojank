package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jwtIssueName = "JWT.ISSUE"

// JWTIssue is the JWT.ISSUE.{userID} privilege.
type JWTIssue struct {
	UserID string
}

func parseJWTIssue(args []string) (Privilege, error) {
	if len(args) != 1 || args[0] == "" {
		return nil, fmt.Errorf("privilege %s requires a user ID", jwtIssueName)
	}
	return JWTIssue{
		UserID: args[0],
	}, nil
}

func (p JWTIssue) String() string {
	return jwtIssueName + "." + p.UserID
}

func (p JWTIssue) Subject() (string, error) {
	if p.UserID == "" {
		return "", fmt.Errorf("privilege %s requires a user ID", jwtIssueName)
	}
	return protodaemon.IssueJWTSubject(p.UserID), nil
}
