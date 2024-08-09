package worker

import (
	"context"
	"github.com/foojank/foojank/internal/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/risor-io/risor"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"
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
	stdin  risoros.File
	stdout risoros.File
}

func New(args Arguments) *Service {
	return &Service{
		args:   args,
		stdin:  NewFile(),
		stdout: NewFile(),
	}
}

func (s *Service) Start(ctx context.Context) error {
	defer func() {
		r := recover()
		// This must not be used in select along with ctx.
		// Doing so would prevent sending the event up to the scheduler, ultimately causing a deadlock.
		s.args.EventCh <- EventWorkerStopped{
			WorkerID: s.args.ID,
			Reason:   r,
		}
	}()

	ros := NewOS(ctx, s.stdin, s.stdout)
	ctx = risoros.WithOS(ctx, ros)

	stdoutSubject := nats.NewInbox()
	ms, err := micro.AddService(s.args.Connection, micro.Config{
		Name:    s.args.Name,
		Version: s.args.Version,
		Metadata: map[string]string{
			"stdout": stdoutSubject,
		},
	})
	if err != nil {
		return err
	}
	defer ms.Stop()

	dataSubject := nats.NewInbox()
	err = ms.AddEndpoint("data", micro.ContextHandler(ctx, s.handleData), micro.WithEndpointSubject(dataSubject))
	if err != nil {
		return err
	}

	stdinSubject := nats.NewInbox()
	err = ms.AddEndpoint("stdin", micro.ContextHandler(ctx, s.handleStdin), micro.WithEndpointSubject(stdinSubject))
	if err != nil {
		return err
	}

	group, groupCtx := errgroup.WithContext(ctx)
	// TODO: rewrite as a service!
	group.Go(func() error {
		for groupCtx.Err() == nil {
			b := make([]byte, 2048)
			_, err := s.stdout.Read(b)
			if err != nil {
				log.Debug("cannot read from stdout")
				continue
			}

			msg := &nats.Msg{
				Subject: stdoutSubject,
				Data:    b,
			}
			err = s.args.Connection.PublishMsg(msg)
			if err != nil {
				log.Debug("cannot publish to stdout subject")
				continue
			}
		}
		return nil
	})

	select {
	case s.args.EventCh <- EventWorkerStarted{
		WorkerID:  s.args.ID,
		ServiceID: ms.Info().ID,
	}:
	case <-ctx.Done():
		return nil
	}

	return group.Wait()
}

func (s *Service) handleData(ctx context.Context, req micro.Request) {
	_ = req.Respond(nil)
	src := string(req.Data())
	log.Debug("before eval")
	_, err := risor.Eval(ctx, src)
	log.Debug("after eval")
	if err != nil {
		_, _ = s.stdout.Write([]byte(err.Error()))
		return
	}
}

func (s *Service) handleStdin(ctx context.Context, req micro.Request) {
	_ = req.Respond(nil)
	log.Debug("before input: %q", string(req.Data()))
	_, _ = s.stdin.Write(req.Data())
	log.Debug("after input")
}
