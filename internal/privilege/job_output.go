package privilege

import (
	"fmt"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const jobOutputName = "JOB.OUTPUT"

// JobOutput is the JOB.OUTPUT privilege.
type JobOutput struct{}

func parseJobOutput(args []string) (Privilege, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("privilege %s does not take arguments", jobOutputName)
	}
	return JobOutput{}, nil
}

func (JobOutput) String() string {
	return jobOutputName
}

func (JobOutput) Subject() (string, error) {
	return protodaemon.GetJobOutputSubject(), nil
}
