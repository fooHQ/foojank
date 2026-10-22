package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jobListName = "JOB.LIST"

// JobList is the JOB.LIST privilege.
type JobList struct{}

func parseJobList(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", jobListName)
	}
	return JobList{}, nil
}

func (JobList) String() string {
	return jobListName
}

func (JobList) Subject() (string, error) {
	return protodaemon.ListJobsSubject(), nil
}
