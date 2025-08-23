package workmanager

import (
	"context"
	"errors"
	"sync"

	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/foojank/internal/router"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/message"
	"github.com/foohq/foojank/internal/vessel/subjects"
	"github.com/foohq/foojank/internal/vessel/worker"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	ID          string
	Templates   subjects.Templates
	Filesystems map[string]risoros.FS
	InputCh     <-chan message.Msg
	OutputCh    chan<- message.Msg
}

type Service struct {
	args    Arguments
	wg      sync.WaitGroup
	workers map[string]state
	eventCh chan message.Msg
}

func New(args Arguments) *Service {
	return &Service{
		args:    args,
		workers: make(map[string]state),
		eventCh: make(chan message.Msg),
	}
}

func (s *Service) Start(ctx context.Context) error {
	api := router.Handlers{
		s.args.Templates.Render(subjects.StartWorker, "<agent>", "<worker>"):      s.handleStartWorker,
		s.args.Templates.Render(subjects.StopWorker, "<agent>", "<worker>"):       s.handleStopWorker,
		s.args.Templates.Render(subjects.WorkerWriteStdin, "<agent>", "<worker>"): s.handleWriteWorkerStdin,
	}
	// Worker event handlers. The keys are not mapped to NATS subjects!
	events := router.Handlers{
		"_WORKER.EVENTS.STOPPED": s.handleWorkerStatusStopped,
		"_WORKER.EVENTS.STDOUT":  s.handleWorkerStatusStdout,
	}

loop:
	for {
		select {
		case msg := <-s.args.InputCh:
			handler, params, ok := api.Match(msg.Subject())
			if !ok {
				_ = msg.Ack()
				continue
			}

			resp := handler(ctx, params, msg.Data())
			if resp == nil {
				_ = msg.Ack()
				continue
			}

			s.sendMessage(ctx, Message{
				msg:     msg,
				subject: s.args.Templates.Render(subjects.Reply, s.args.ID, msg.ID()),
				data:    resp,
			})

		case msg := <-s.eventCh:
			handler, params, ok := events.Match(msg.Subject())
			if !ok {
				continue
			}

			resp := handler(ctx, params, msg.Data())
			if resp == nil {
				continue
			}

			s.sendMessage(ctx, resp.(message.Msg))

		case <-ctx.Done():
			break loop
		}
	}

	// Close worker contexts.
	for id := range s.workers {
		_ = s.stopWorker(id)
	}

	// Process remaining events from workers.
	// When all events were processed, the worker pool will be empty.
	for len(s.workers) > 0 {
		msg := <-s.eventCh
		handler, params, ok := events.Match(msg.Subject())
		if !ok {
			continue
		}

		resp := handler(ctx, params, msg.Data())
		if resp == nil {
			continue
		}

		// TODO: make sure EventWorkerStopped is forwarded to encoder | publisher
		// 	This requires changing shutdown logic of the services so that encoder | publisher stop as last.
		s.sendMessage(ctx, resp.(message.Msg))
	}

	s.wg.Wait()
	return nil
}

func (s *Service) handleStartWorker(ctx context.Context, params router.Params, data any) any {
	workerID, ok := params["worker"]
	if !ok {
		log.Debug("Missing worker ID")
		return proto.StartWorkerResponse{
			Error: errors.New("missing worker id"),
		}
	}

	v, ok := data.(proto.StartWorkerRequest)
	if !ok {
		log.Debug("Invalid request data")
		return proto.StartWorkerResponse{
			Error: errors.New("invalid request data"),
		}
	}

	w, err := s.startWorker(workerID, v.File, v.Args, v.Env)
	if err != nil {
		log.Debug("Cannot start worker", "error", err)
		return proto.StartWorkerResponse{
			Error: err,
		}
	}

	s.addWorker(workerID, w)

	return proto.StartWorkerResponse{}
}

func (s *Service) handleStopWorker(ctx context.Context, params router.Params, data any) any {
	workerID, ok := params["worker"]
	if !ok {
		log.Debug("Missing worker ID")
		return proto.StopWorkerResponse{
			Error: errors.New("missing worker id"),
		}
	}

	err := s.stopWorker(workerID)
	if err != nil {
		log.Debug("Cannot stop worker", "error", err)
		return proto.StopWorkerResponse{
			Error: err,
		}
	}

	return proto.StopWorkerResponse{}
}

func (s *Service) handleWriteWorkerStdin(ctx context.Context, params router.Params, data any) any {
	workerID, ok := params["worker"]
	if !ok {
		log.Debug("Missing worker ID")
		return nil
	}

	v, ok := data.(proto.UpdateWorkerStdio)
	if !ok {
		log.Debug("Invalid request data")
		return nil
	}

	err := s.writeWorkerStdin(ctx, workerID, v.Data)
	if err != nil {
		log.Debug("Cannot write worker stdin", "error", err)
		return nil
	}

	return nil
}

func (s *Service) handleWorkerStatusStopped(_ context.Context, params router.Params, data any) any {
	v, ok := data.(worker.EventWorkerStopped)
	if !ok {
		log.Debug("Invalid event data")
		return nil
	}

	s.removeWorker(v.WorkerID)

	return Message{
		subject: s.args.Templates.Render(subjects.WorkerStatus, s.args.ID, v.WorkerID),
		data: proto.UpdateWorkerStatus{
			Status: int64(v.Status),
		},
	}
}

func (s *Service) handleWorkerStatusStdout(_ context.Context, _ router.Params, data any) any {
	v, ok := data.(worker.EventWorkerOutput)
	if !ok {
		log.Debug("Invalid event data")
		return nil
	}

	return Message{
		subject: s.args.Templates.Render(subjects.WorkerWriteStdout, s.args.ID, v.WorkerID),
		data: proto.UpdateWorkerStdio{
			Data: v.OutputData,
		},
	}
}

type state struct {
	stdinCh chan []byte
	cancel  context.CancelFunc
}

func (s *Service) startWorker(id, file string, args, env []string) (state, error) {
	_, ok := s.workers[id]
	if ok {
		return state{}, errors.New("worker already exists")
	}

	stdinCh := make(chan []byte)
	wCtx, cancel := context.WithCancel(context.Background())
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := worker.New(worker.Arguments{
			ID:          id,
			File:        file,
			Args:        args,
			Env:         env,
			EventCh:     s.eventCh,
			StdinCh:     stdinCh,
			Filesystems: s.args.Filesystems,
		}).Start(wCtx)
		if err != nil {
			log.Debug("Worker stopped with an error", "error", err)
		}
	}()
	return state{
		stdinCh: stdinCh,
		cancel:  cancel,
	}, nil
}

func (s *Service) stopWorker(id string) error {
	w, ok := s.workers[id]
	if !ok {
		return errors.New("worker not found")
	}
	w.cancel()
	return nil
}

func (s *Service) addWorker(id string, w state) {
	s.workers[id] = w
}

func (s *Service) removeWorker(id string) {
	delete(s.workers, id)
}

func (s *Service) writeWorkerStdin(ctx context.Context, id string, data []byte) error {
	w, ok := s.workers[id]
	if !ok {
		return errors.New("worker not found")
	}
	select {
	case w.stdinCh <- data:
	case <-ctx.Done():
	}
	return nil
}

func (s *Service) sendMessage(ctx context.Context, msg message.Msg) {
	select {
	case s.args.OutputCh <- msg:
	case <-ctx.Done():
	}
}

type Message struct {
	msg     message.Msg
	subject string
	data    any
}

func (m Message) ID() string {
	return ""
}

func (m Message) Subject() string {
	return m.subject
}

func (m Message) Data() any {
	return m.data
}

func (m Message) Ack() error {
	if m.msg == nil {
		return nil
	}
	return m.msg.Ack()
}
