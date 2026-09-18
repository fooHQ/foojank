package privilege

import (
	"fmt"
	"strings"
)

// Privilege is a user capability that expands to a NATS subject.
type Privilege interface {
	String() string
	Subject() (string, error)
}

// parser reconstructs a Privilege from the arguments that follow the kind
// name. Arguments are the remaining dotted segments; a nil slice means the
// stored name was exactly the kind.
type parser func(args []string) (Privilege, error)

var parsers = map[string]parser{
	userCreateName: parseUserCreate,
	userGetName:    parseUserGet,
	userListName:   parseUserList,
	jwtIssueName:   parseJWTIssue,
}

// Privileges is a parsed set of privileges and the NATS permissions they grant.
type Privileges struct {
	privileges  []Privilege
	permissions []string
}

// Privileges returns the parsed privileges. An empty set is returned as nil.
func (p Privileges) Privileges() []Privilege {
	return p.privileges
}

// Permissions returns the NATS subjects granted by the privileges. An empty
// set is returned as nil.
func (p Privileges) Permissions() []string {
	return p.permissions
}

func parsePrivilege(s string) (Privilege, error) {
	upper := strings.ToUpper(s)
	if parse, ok := parsers[upper]; ok {
		return parse(nil)
	}

	rest := upper
	for {
		i := strings.LastIndexByte(rest, '.')
		if i < 0 {
			break
		}

		if parse, ok := parsers[rest[:i]]; ok {
			return parse(strings.Split(s[i+1:], "."))
		}
		rest = rest[:i]
	}

	return nil, fmt.Errorf("unknown privilege %q", s)
}

// ParsePrivileges parses stored privilege names. An empty slice is returned as
// a zero Privileges value.
func ParsePrivileges(ss []string) (Privileges, error) {
	if len(ss) == 0 {
		return Privileges{}, nil
	}

	privileges := make([]Privilege, len(ss))
	permissions := make([]string, len(ss))
	for i, s := range ss {
		p, err := parsePrivilege(s)
		if err != nil {
			return Privileges{}, err
		}
		subj, err := p.Subject()
		if err != nil {
			return Privileges{}, err
		}
		privileges[i] = p
		permissions[i] = subj
	}
	return Privileges{
		privileges:  privileges,
		permissions: permissions,
	}, nil
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
