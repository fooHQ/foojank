package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const auditListName = "AUDIT.LIST"

// AuditList is the AUDIT.LIST privilege.
type AuditList struct{}

func parseAuditList(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", auditListName)
	}
	return AuditList{}, nil
}

func (AuditList) String() string {
	return auditListName
}

func (AuditList) Subject() (string, error) {
	return protodaemon.ListAuditSubject(), nil
}
