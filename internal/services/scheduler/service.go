package scheduler

import (
	"context"
	"github.com/foojank/foojank/internal/log"
	"github.com/foojank/foojank/internal/services/decoder"
	"github.com/foojank/foojank/internal/services/worker"
	"github.com/nats-io/nats.go"
	"sync"
)

type Arguments struct {
	Connection *nats.Conn
	InputCh    <-chan decoder.Message
}

type Service struct {
	args Arguments
	wg   sync.WaitGroup
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	var workerID uint64
	var workers = make(map[uint64]state)
	var eventCh = make(chan worker.Event)

loop:
	for {
		select {
		case msg := <-s.args.InputCh:
			data := msg.Data()
			switch v := data.(type) {
			case decoder.CreateWorkerRequest:
				workerID++
				wCtx, cancel := context.WithCancel(ctx)
				workers[workerID] = state{
					w:      s.createWorker(wCtx, workerID, eventCh),
					cancel: cancel,
				}

				_ = msg.Reply(decoder.CreateWorkerResponse{
					ID: workerID,
				})

			case decoder.DestroyWorkerRequest:
				_, ok := workers[v.ID]
				if !ok {
					_ = msg.ReplyError("400", "worker does not exist", nil)
					continue
				}

				workers[v.ID].Cancel()
				_ = msg.Reply(decoder.DestroyWorkerResponse{})

			case decoder.GetWorkerRequest:
				_, ok := workers[v.ID]
				if !ok {
					_ = msg.ReplyError("400", "worker does not exist", nil)
					continue
				}

				_ = msg.Reply(decoder.GetWorkerResponse{
					ServiceID: workers[v.ID].ServiceID(),
				})
			}

		case event := <-eventCh:
			switch v := event.(type) {
			case worker.EventWorkerStarted:
				log.Debug("%#v", v)
				workers[v.WorkerID] = state{
					w:         workers[v.WorkerID].w,
					serviceID: v.ServiceID,
					cancel:    workers[v.WorkerID].cancel,
				}

			case worker.EventWorkerStopped:
				log.Debug("%#v", v)
				workers[v.WorkerID].Cancel()
				delete(workers, v.WorkerID)
			}

		case <-ctx.Done():
			break loop
		}
	}

	log.Debug("cancelling all running workers")
	for i := range workers {
		log.Debug("worker id=%d cancelled", i)
		workers[i].Cancel()
		<-eventCh
	}

	log.Debug("waiting for all workers to stop")
	s.wg.Wait()
	return nil
}

func (s *Service) createWorker(ctx context.Context, workerID uint64, eventCh chan<- worker.Event) *worker.Service {
	log.Debug("creating a new worker id=%d", workerID)
	w := worker.New(worker.Arguments{
		Connection: s.args.Connection,
		EventCh:    eventCh,
		ID:         workerID,
		Name:       "vessel-worker", // TODO: configurable
		Version:    "0.1.0",         // TODO: configurable
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := w.Start(ctx)
		if err != nil {
			log.Debug("worker stopped with an error: %v", err)
			return
		}
	}()

	return w
}

type state struct {
	w         *worker.Service
	serviceID string
	cancel    context.CancelFunc
}

func (s state) ServiceID() string {
	return s.serviceID
}

func (s state) Cancel() {
	s.cancel()
}
