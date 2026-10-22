package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jobCreateName = "JOB.CREATE"

// JobCreate is the JOB.CREATE privilege.
type JobCreate struct{}

func parseJobCreate(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", jobCreateName)
	}
	return JobCreate{}, nil
}

func (JobCreate) String() string {
	return jobCreateName
}

func (JobCreate) Subject() (string, error) {
	return protodaemon.CreateJobSubject(), nil
}
