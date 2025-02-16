package worker

import (
	"context"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/connector"
	"github.com/foohq/foojank/internal/vessel/worker/decoder"
	"github.com/foohq/foojank/internal/vessel/worker/processor"
	"github.com/foohq/foojank/internal/vessel/worker/publisher"
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
		if r != nil {
			// This must not be used in select along with ctx.
			// Doing so would prevent sending the event up to the scheduler, ultimately causing a deadlock.
			s.args.EventCh <- EventWorkerStopped{
				WorkerID: s.args.ID,
				Reason:   r,
			}
		}
	}()

	jetStream, err := jetstream.New(s.args.Connection)
	if err != nil {
		log.Debug("cannot create a JetStream context", "error", err.Error())
		return err
	}

	dataSubject := s.args.Name + "." + strconv.FormatUint(s.args.ID, 10) + ".DATA"
	stdinSubject := s.args.Name + "." + strconv.FormatUint(s.args.ID, 10) + ".STDIN"
	stdoutSubject := s.args.Name + "." + strconv.FormatUint(s.args.ID, 10) + ".STDOUT"

	connectorInfoCh := make(chan connector.InfoMessage, 1)
	connectorOutCh := make(chan connector.Message)
	decoderDataCh := make(chan decoder.Message)
	decoderStdinCh := make(chan decoder.Message)
	processorStdoutCh := make(chan []byte, 1024)

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
			InfoCh:       connectorInfoCh,
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
			DataCh:    decoderDataCh,
			StdinCh:   decoderStdinCh,
			StdoutCh:  processorStdoutCh,
			JetStream: jetStream,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return publisher.New(publisher.Arguments{
			Connection: s.args.Connection,
			Subject:    stdoutSubject,
			InputCh:    processorStdoutCh,
		}).Start(groupCtx)
	})

	select {
	case info := <-connectorInfoCh:
		select {
		case s.args.EventCh <- EventWorkerStarted{
			WorkerID:    s.args.ID,
			ServiceName: info.ServiceName(),
			ServiceID:   info.ServiceID(),
		}:
		case <-ctx.Done():
			break
		}
	case <-ctx.Done():
		break
	}

	err = group.Wait()
	// This must not be used in select along with ctx.
	// Doing so would prevent sending the event up to the scheduler, ultimately causing a deadlock.
	s.args.EventCh <- EventWorkerStopped{
		WorkerID: s.args.ID,
		Reason:   err,
	}

	return nil
}
