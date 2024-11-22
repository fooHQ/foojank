package scheduler

import (
	"context"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/decoder"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	worker2 "github.com/foohq/foojank/internal/vessel/worker"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"sync"
)

type Arguments struct {
	Connection *nats.Conn
	Stream     jetstream.JetStream
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
	var eventCh = make(chan worker2.Event)

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
					_ = msg.ReplyError(errcodes.ErrWorkerNotFound, "worker does not exist", nil)
					continue
				}

				workers[v.ID].Cancel()
				_ = msg.Reply(decoder.DestroyWorkerResponse{})

			case decoder.GetWorkerRequest:
				w, ok := workers[v.ID]
				if !ok {
					_ = msg.ReplyError(errcodes.ErrWorkerNotFound, "worker does not exist", nil)
					continue
				}

				if w.ServiceID() == "" {
					_ = msg.ReplyError(errcodes.ErrWorkerStarting, "worker is starting", nil)
					continue
				}

				_ = msg.Reply(decoder.GetWorkerResponse{
					ServiceName: w.ServiceName(),
					ServiceID:   w.ServiceID(),
				})
			}

		case event := <-eventCh:
			switch v := event.(type) {
			case worker2.EventWorkerStarted:
				log.Debug("%#v", v)
				workers[v.WorkerID] = state{
					w:           workers[v.WorkerID].w,
					serviceName: v.ServiceName,
					serviceID:   v.ServiceID,
					cancel:      workers[v.WorkerID].cancel,
				}

			case worker2.EventWorkerStopped:
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

func (s *Service) createWorker(ctx context.Context, workerID uint64, eventCh chan<- worker2.Event) *worker2.Service {
	log.Debug("creating a new worker id=%d", workerID)
	w := worker2.New(worker2.Arguments{
		ID:         workerID,
		Name:       config.ServiceWorkerName,
		Version:    config.ServiceVersion,
		Connection: s.args.Connection,
		Stream:     s.args.Stream,
		EventCh:    eventCh,
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
	w           *worker2.Service
	serviceName string
	serviceID   string
	cancel      context.CancelFunc
}

func (s state) ServiceName() string {
	return s.serviceName
}

func (s state) ServiceID() string {
	return s.serviceID
}

func (s state) Cancel() {
	s.cancel()
}
