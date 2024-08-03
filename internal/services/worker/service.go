package worker

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/risor-io/risor"
	risoros "github.com/risor-io/risor/os"
)

type Arguments struct {
	Connection *nats.Conn
	EventCh    chan<- Event
	ID         uint64
	Name       string
	Version    string
}

type Service struct {
	args   Arguments
	stdout *risoros.BufferFile
	stdin  *risoros.BufferFile
}

func New(args Arguments) *Service {
	return &Service{
		args:   args,
		stdout: risoros.NewBufferFile(nil),
		stdin:  risoros.NewBufferFile(nil), // TODO: this is actually not thread-safe!!!
	}
}

func (s *Service) Start(ctx context.Context) error {
	defer func() {
		r := recover()
		s.args.EventCh <- EventWorkerStopped{
			WorkerID: s.args.ID,
			Reason:   r,
		}
	}()

	ros := NewOS(ctx, s.stdout) // TODO: add stdin
	ctx = risoros.WithOS(ctx, ros)

	ms, err := micro.AddService(s.args.Connection, micro.Config{
		Name:    s.args.Name,
		Version: s.args.Version,
	})
	if err != nil {
		return err
	}
	defer ms.Stop()

	dataSubject := nats.NewInbox()
	stdinSubject := nats.NewInbox()
	err = ms.AddEndpoint("data", micro.ContextHandler(ctx, s.handleData), micro.WithEndpointSubject(dataSubject))
	if err != nil {
		return err
	}
	err = ms.AddEndpoint("stdin", micro.ContextHandler(ctx, s.handleStdin), micro.WithEndpointSubject(stdinSubject))
	if err != nil {
		return err
	}

	s.args.EventCh <- EventWorkerStarted{
		WorkerID:  s.args.ID,
		ServiceID: ms.Info().ID,
	}

	<-ctx.Done()
	return nil
}

func (s *Service) handleData(ctx context.Context, req micro.Request) {
	src := string(req.Data())
	_, err := risor.Eval(ctx, src)
	if err != nil {
		_ = req.Error("400", err.Error(), nil)
		return
	}

	_ = req.Respond(s.stdout.Bytes())
}

func (s *Service) handleStdin(ctx context.Context, req micro.Request) {
	// TODO
}
