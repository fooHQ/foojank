package dispatcher

import (
	"context"
	"errors"
	"sync"

	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/vessel/connector"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	InputCh       <-chan connector.Message
	SubjectPrefix string
	ServiceID     string
}

type Service struct {
	args    Arguments
	workers map[string]*state
	wg      sync.WaitGroup
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	for {
		select {
		case msg := <-s.args.InputCh:
			err := s.handleMessage(ctx, msg)
			if err != nil {
				log.Debug("cannot handle message", "error", err)
				continue
			}

		// TODO: monitor worker events!

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) handleMessage(ctx context.Context, msg connector.Message) error {
	// Acknowledge a message.
	defer msg.Ack()
	switch v := msg.Data().(type) {
	case proto.CreateJobRequest:
		reply := s.createJob(ctx, v)
		return msg.Reply(ctx, reply)

	case proto.CancelJobRequest:
		reply := s.cancelJob(ctx, v)
		return msg.Reply(ctx, reply)

	default:
		return errors.New("unsupported message type")
	}
}

func (s *Service) createJob(ctx context.Context, req proto.CreateJobRequest) proto.CreateJobResponse {
	jobID := nuid.Next()
	worker := startWorker(ctx, &s.wg)
	s.workers[jobID] = worker
	return proto.CreateJobResponse{
		JobID:         jobID,
		StdinSubject:  s.args.SubjectPrefix + "." + s.args.ServiceID + "." + jobID + ".stdin",
		StdoutSubject: s.args.SubjectPrefix + "." + s.args.ServiceID + "." + jobID + ".stdout",
	}
}

func (s *Service) cancelJob(ctx context.Context, req proto.CancelJobRequest) proto.CancelJobResponse {
	return proto.CancelJobResponse{}
}

type state struct {
	// TODO
	cancel context.CancelFunc
}

func startWorker(ctx context.Context, wg *sync.WaitGroup) *state {
	wCtx, cancel := context.WithCancel(ctx)
	_ = wCtx // TODO: can be removed!
	s := &state{
		cancel: cancel,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: start worker here
	}()
	return s
}
