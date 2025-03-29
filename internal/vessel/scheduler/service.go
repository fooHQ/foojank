package scheduler

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	filefs "github.com/foohq/foojank/filesystems/file"
	natsfs "github.com/foohq/foojank/filesystems/nats"
	engineos "github.com/foohq/foojank/internal/engine/os"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/decoder"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker"
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

	jetStream, err := jetstream.New(s.args.Connection)
	if err != nil {
		log.Debug("cannot create a JetStream context", "error", err.Error())
		return err
	}

	store, err := jetStream.ObjectStore(ctx, config.ServiceName)
	if err != nil {
		log.Debug("cannot open object store", "error", err)
		return err
	}

	uriHandlers := make(map[string]engineos.URIHandler)
	uriHandlers["file"], err = filefs.NewURIHandler()
	if err != nil {
		log.Debug("cannot create file handler", "error", err)
		return err
	}

	uriHandlers["nats"], err = natsfs.NewURIHandler(ctx, store)
	if err != nil {
		log.Debug("cannot create nats handler", "error", err)
		return err
	}

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
					w:      s.createWorker(wCtx, workerID, uriHandlers, eventCh),
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
			case worker.EventWorkerStarted:
				log.Debug("received worker started event", "event", v)
				workers[v.WorkerID] = state{
					w:           workers[v.WorkerID].w,
					serviceName: v.ServiceName,
					serviceID:   v.ServiceID,
					cancel:      workers[v.WorkerID].cancel,
				}

			case worker.EventWorkerStopped:
				log.Debug("received worker stopped event", "event", v)
				workers[v.WorkerID].Cancel()
				delete(workers, v.WorkerID)
			}

		case <-ctx.Done():
			break loop
		}
	}

	log.Debug("cancelling all running workers")
	for i := range workers {
		log.Debug("worker cancelled", "id", i)
		workers[i].Cancel()
		<-eventCh
	}

	log.Debug("waiting for all workers to stop")
	s.wg.Wait()
	log.Debug("all workers stopped")
	return nil
}

func (s *Service) createWorker(
	ctx context.Context,
	workerID uint64,
	uriHandlers map[string]engineos.URIHandler,
	eventCh chan<- worker.Event,
) *worker.Service {
	log.Debug("creating a new worker", "id", workerID)
	w := worker.New(worker.Arguments{
		ID:          workerID,
		Name:        config.ServiceName,
		Version:     config.ServiceVersion,
		Connection:  s.args.Connection,
		URIHandlers: uriHandlers,
		EventCh:     eventCh,
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := w.Start(ctx)
		if err != nil {
			log.Debug("worker stopped", "error", err)
			return
		}
	}()

	return w
}

type state struct {
	w           *worker.Service
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
