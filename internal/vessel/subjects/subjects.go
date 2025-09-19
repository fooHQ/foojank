package subjects

import "fmt"

const (
	StartWorker = iota
	StopWorker
	WorkerWriteStdin
	WorkerWriteStdout
	WorkerStatus
	ConnInfo
	Reply
)

type Templates map[int]string

func (t Templates) Render(subject int, arg ...any) string {
	return fmt.Sprintf(t[subject], arg...)
}
