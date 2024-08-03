package worker

type Event any

type EventWorkerStarted struct {
	WorkerID  uint64
	ServiceID string
}

type EventWorkerStopped struct {
	WorkerID uint64
	Reason   any
}
