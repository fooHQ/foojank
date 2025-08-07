package dispatcher

import (
	"context"
	"errors"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"
	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/foojank/internal/vessel/connector"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker2"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	Connection jetstream.JetStream
	Stream     string
	InputCh    <-chan connector.Message
}

type Service struct {
	args        Arguments
	workers     map[string]*state
	wg          sync.WaitGroup
	filesystems map[string]risoros.FS
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
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
				return errors.New("unsupported message type")
			}

			err := msg.Reply(ctx, reply)
			if err != nil {
				log.Debug("cannot send reply")
				continue loop
			}

			err = msg.Ack()
			if err != nil {
				log.Debug("cannot ack message")
				continue loop
			}

		// TODO: monitor worker events!

		case <-ctx.Done():
			break loop
		}
	}

	for _, w := range s.workers {
		w.cancel()
		// TODO Wait for event in eventCh
	}

	s.wg.Wait()
	return nil
}

func (s *Service) createJob(ctx context.Context, req proto.CreateJobRequest) proto.CreateJobResponse {
	log.Debug("create job", "command", req.Command, "args", req.Args, "env", req.Env)
	id := nuid.Next()
	s.startWorker(ctx, id, req.Command, req.Args, req.Env)
	return proto.CreateJobResponse{
		JobID: id,
	}
}

func (s *Service) cancelJob(ctx context.Context, req proto.CancelJobRequest) proto.CancelJobResponse {
	log.Debug("cancel job", "id", req.JobID)
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
) {
	wCtx, cancel := context.WithCancel(ctx)
	stdin := ""  // TODO
	stdout := "" // TODO
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := worker.New(worker.Arguments{
			ID:            id,
			Stream:        s.args.Stream,
			StdinSubject:  stdin,
			StdoutSubject: stdout,
			Entrypoint:    entrypoint,
			Args:          args,
			Env:           env,
			Connection:    s.args.Connection,
			Filesystems:   s.filesystems,
			EventCh:       nil, // TODO
		}).Start(wCtx)
		if err != nil {
			log.Debug("worker exited with error", "error", err)
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
