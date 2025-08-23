package worker

import "github.com/foohq/foojank/internal/vessel/message"

type EventWorkerStarted struct {
	WorkerID string
}

func (e EventWorkerStarted) ID() string {
	return ""
}

func (e EventWorkerStarted) Subject() string {
	return "_WORKER.EVENTS.STARTED"
}

func (e EventWorkerStarted) Data() any {
	return e
}

func (e EventWorkerStarted) Ack() error {
	return message.ErrUnsupported
}

type EventWorkerOutput struct {
	WorkerID   string
	OutputData []byte
}

func (e EventWorkerOutput) ID() string {
	return ""
}

func (e EventWorkerOutput) Subject() string {
	return "_WORKER.EVENTS.STDOUT"
}

func (e EventWorkerOutput) Data() any {
	return e
}

func (e EventWorkerOutput) Ack() error {
	return message.ErrUnsupported
}

type EventWorkerStopped struct {
	WorkerID string
	Status   int
	Error    error
}

func (e EventWorkerStopped) ID() string {
	return ""
}

func (e EventWorkerStopped) Subject() string {
	return "_WORKER.EVENTS.STOPPED"
}

func (e EventWorkerStopped) Data() any {
	return e
}

func (e EventWorkerStopped) Ack() error {
	return message.ErrUnsupported
}
