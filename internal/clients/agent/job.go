package agent

const (
	JobStatusPending    = "Pending"
	JobStatusRunning    = "Running"
	JobStatusCancelling = "Cancelling"
	JobStatusCancelled  = "Cancelled"
	JobStatusFinished   = "Finished"
	JobStatusFailed     = "Failed"
)

type Job struct {
	ID      string
	AgentID string
	Command string
	Args    string
	Status  string
	Error   error
}
