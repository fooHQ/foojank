package auth

import (
	"fmt"
	"strings"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jwtIssuePrefix = "JWT.ISSUE."

// Privilege is a user capability that expands to a NATS subject.
type Privilege interface {
	String() string
	Subject() (string, error)
}

// UserCreate is the USER.CREATE privilege.
type UserCreate struct{}

func (UserCreate) String() string {
	return "USER.CREATE"
}

func (UserCreate) Subject() (string, error) {
	return protodaemon.CreateUserSubject(), nil
}

// UserGet is the USER.GET privilege.
type UserGet struct{}

func (UserGet) String() string {
	return "USER.GET"
}

func (UserGet) Subject() (string, error) {
	return protodaemon.GetUserSubject(), nil
}

// UserList is the USER.LIST privilege.
type UserList struct{}

func (UserList) String() string {
	return "USER.LIST"
}

func (UserList) Subject() (string, error) {
	return protodaemon.ListUsersSubject(), nil
}

// JWTIssue is the JWT.ISSUE.{userID} privilege.
type JWTIssue struct {
	UserID string
}

func (p JWTIssue) String() string {
	return jwtIssuePrefix + p.UserID
}

func (p JWTIssue) Subject() (string, error) {
	if p.UserID == "" {
		return "", fmt.Errorf("privilege JWT.ISSUE requires a user ID")
	}
	return protodaemon.IssueJWTSubject(p.UserID), nil
}

// ParsePrivilege parses a stored privilege name.
func ParsePrivilege(s string) (Privilege, error) {
	upper := strings.ToUpper(s)
	switch upper {
	case UserCreate{}.String():
		return UserCreate{}, nil
	case UserGet{}.String():
		return UserGet{}, nil
	case UserList{}.String():
		return UserList{}, nil
	}

	if strings.HasPrefix(upper, jwtIssuePrefix) {
		userID := s[len(jwtIssuePrefix):]
		if userID == "" {
			return nil, fmt.Errorf("privilege JWT.ISSUE requires a user ID")
		}
		return JWTIssue{UserID: userID}, nil
	}
	if upper == "JWT.ISSUE" {
		return nil, fmt.Errorf("privilege JWT.ISSUE requires a user ID")
	}

	return nil, fmt.Errorf("unknown privilege %q", s)
}

// ParsePrivileges parses stored privilege names. An empty slice is returned as nil.
func ParsePrivileges(ss []string) ([]Privilege, error) {
	if len(ss) == 0 {
		return nil, nil
	}

	privileges := make([]Privilege, len(ss))
	for i, s := range ss {
		p, err := ParsePrivilege(s)
		if err != nil {
			return nil, err
		}
		privileges[i] = p
	}
	return privileges, nil
}

// FormatPrivileges encodes privileges for storage. An empty slice is returned as nil.
func FormatPrivileges(privileges []Privilege) []string {
	if len(privileges) == 0 {
		return nil
	}

	ss := make([]string, len(privileges))
	for i, p := range privileges {
		ss[i] = p.String()
	}
	return ss
}
