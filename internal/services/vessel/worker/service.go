package worker

import (
	"context"
	"github.com/foojank/foojank/internal/services/vessel/worker/connector"
	"github.com/foojank/foojank/internal/services/vessel/worker/decoder"
	"github.com/foojank/foojank/internal/services/vessel/worker/processor"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	ID         uint64
	Name       string
	Version    string
	Connection *nats.Conn
	EventCh    chan<- Event
}

type Service struct {
	args Arguments
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
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

	// TODO: allow these to be configured from the outside
	dataSubject := nats.NewInbox()
	stdinSubject := nats.NewInbox()
	stdoutSubject := nats.NewInbox()

	connectorIDCh := make(chan string, 1)
	connectorOutCh := make(chan connector.Message, 65535)
	decoderDataCh := make(chan decoder.Message, 65535)
	decoderStdinCh := make(chan decoder.Message, 65535)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return connector.New(connector.Arguments{
			Name:    s.args.Name,
			Version: s.args.Version,
			Metadata: map[string]string{
				// TODO: progname
				// TODO: args
				// TODO: environ?
				"stdout": stdoutSubject,
			},
			StdinSubject: stdinSubject,
			DataSubject:  dataSubject,
			Connection:   s.args.Connection,
			IDCh:         connectorIDCh,
			OutputCh:     connectorOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return decoder.New(decoder.Arguments{
			DataSubject: dataSubject,
			InputCh:     connectorOutCh,
			DataCh:      decoderDataCh,
			StdinCh:     decoderStdinCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return processor.New(processor.Arguments{
			Connection:    s.args.Connection,
			StdoutSubject: stdoutSubject,
			DataCh:        decoderDataCh,
			StdinCh:       decoderStdinCh,
		}).Start(groupCtx)
	})

	var id string
	select {
	case v := <-connectorIDCh:
		id = v
	case <-ctx.Done():
		return nil
	}

	select {
	case s.args.EventCh <- EventWorkerStarted{
		WorkerID:  s.args.ID,
		ServiceID: id,
	}:
	case <-ctx.Done():
		return nil
	}

	return group.Wait()
}
