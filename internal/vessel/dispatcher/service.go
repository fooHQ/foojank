package dispatcher

import (
	"context"
	"errors"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"
	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/message"
	"github.com/foohq/foojank/internal/vessel/worker"
	"github.com/foohq/foojank/proto"
)

// TODO: rename package to workmanager

type Arguments struct {
	Connection    jetstream.JetStream
	Stream        string
	StdinSubject  string
	StdoutSubject string
	InputCh       <-chan message.Msg
}

type Service struct {
	args        Arguments
	workers     map[string]*state
	eventCh     chan any
	wg          sync.WaitGroup
	filesystems map[string]risoros.FS
}

func New(args Arguments) *Service {
	return &Service{
		args:    args,
		workers: make(map[string]*state),
		eventCh: make(chan any),
	}
}

func (s *Service) Start(ctx context.Context) error {
	log.Debug("Service started", "service", "vessel.workmanager")
	defer log.Debug("Service stopped", "service", "vessel.workmanager")

loop:
	for {
		select {
		case msg := <-s.args.InputCh:
			var reply any
			switch v := msg.Data().(type) {
			case proto.CreateJobRequest:
				reply = s.createJob(ctx, v)

			case proto.CancelJobRequest:
				reply = s.cancelJob(ctx, v)

			default:
				log.Debug("Unsupported message type", "message", v)
				_ = msg.Ack()
				continue loop
			}

			err := msg.Reply(ctx, reply)
			if err != nil {
				log.Debug("Cannot send a reply")
				continue loop
			}

			err = msg.Ack()
			if err != nil {
				log.Debug("Cannot acknowledge message")
				continue loop
			}

		case msg := <-s.eventCh:
			switch v := msg.(type) {
			case worker.EventWorkerStarted:
				// Consume the message and nothing else
			case worker.EventWorkerStopped:
				_ = s.stopWorker(v.ID)
			default:
				log.Debug("Unsupported event type", "event", v)
				continue loop
			}

		case <-ctx.Done():
			break loop
		}
	}

	for id := range s.workers {
		_ = s.stopWorker(id)
		<-s.eventCh
	}

	s.wg.Wait()
	return nil
}

func (s *Service) createJob(ctx context.Context, req proto.CreateJobRequest) proto.CreateJobResponse {
	id := nuid.Next()
	// TODO: set stdinSubject, stdoutSubject
	s.startWorker(ctx, id, req.Command, req.Args, req.Env, "", "", s.eventCh)
	return proto.CreateJobResponse{
		JobID: id,
	}
}

func (s *Service) cancelJob(ctx context.Context, req proto.CancelJobRequest) proto.CancelJobResponse {
	id := req.JobID
	err := s.stopWorker(id)
	if err != nil {
		return proto.CancelJobResponse{
			Error: err,
		}
	}
	return proto.CancelJobResponse{}
}

type state struct {
	cancel context.CancelFunc
}

func (s *Service) startWorker(
	ctx context.Context,
	id,
	entrypoint string,
	args,
	env []string,
	stdinSubject,
	stdoutSubject string,
	eventCh chan any,
) {
	wCtx, cancel := context.WithCancel(ctx)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := worker.New(worker.Arguments{
			ID:            id,
			Stream:        s.args.Stream,
			StdinSubject:  stdinSubject,
			StdoutSubject: stdoutSubject,
			Entrypoint:    entrypoint,
			Args:          args,
			Env:           env,
			Connection:    s.args.Connection,
			Filesystems:   s.filesystems,
			EventCh:       eventCh,
		}).Start(wCtx)
		if err != nil {
			log.Debug("Worker stopped with an error", "error", err)
		}
	}()
	s.workers[id] = &state{
		cancel: cancel,
	}
}

func (s *Service) stopWorker(id string) error {
	w, ok := s.workers[id]
	if !ok {
		return errors.New("worker not found")
	}
	w.cancel()
	delete(s.workers, id)
	return nil
}
