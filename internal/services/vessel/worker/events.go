package worker

type Event any

type EventWorkerStarted struct {
	WorkerID    uint64
	ServiceName string
	ServiceID   string
}

type EventWorkerStopped struct {
	WorkerID uint64
	Reason   any
}
