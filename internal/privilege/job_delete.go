package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jobDeleteName = "JOB.DELETE"

// JobDelete is the JOB.DELETE privilege.
type JobDelete struct{}

func parseJobDelete(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", jobDeleteName)
	}
	return JobDelete{}, nil
}

func (JobDelete) String() string {
	return jobDeleteName
}

func (JobDelete) Subject() (string, error) {
	return protodaemon.DeleteJobSubject(), nil
}
