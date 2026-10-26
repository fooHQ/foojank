package role

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var ErrParserError = errors.New("parser error")

// Role is one role in an import file. The file is a JSON array of these
// objects. Storage fields such as created_at are ignored.
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Privileges  []string `json:"privileges"`
}

func ParseFile(file string) ([]Role, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var roles []Role
	err = json.Unmarshal(b, &roles)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	seen := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		if role.Name == "" {
			return nil, errors.New("role name is required")
		}
		if _, ok := seen[role.Name]; ok {
			return nil, fmt.Errorf("duplicate role %q", role.Name)
		}
		seen[role.Name] = struct{}{}
	}

	return roles, nil
}
