package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jobCancelName = "JOB.CANCEL"

// JobCancel is the JOB.CANCEL privilege.
type JobCancel struct{}

func parseJobCancel(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", jobCancelName)
	}
	return JobCancel{}, nil
}

func (JobCancel) String() string {
	return jobCancelName
}

func (JobCancel) Subject() (string, error) {
	return protodaemon.CancelJobSubject(), nil
}
