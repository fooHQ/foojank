package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jobGetName = "JOB.GET"

// JobGet is the JOB.GET privilege.
type JobGet struct{}

func parseJobGet(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", jobGetName)
	}
	return JobGet{}, nil
}

func (JobGet) String() string {
	return jobGetName
}

func (JobGet) Subject() (string, error) {
	return protodaemon.GetJobSubject(), nil
}
