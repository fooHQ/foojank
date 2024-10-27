package vessel

import (
	"context"
	"github.com/foojank/foojank/internal/services/vessel/connector"
	"github.com/foojank/foojank/internal/services/vessel/decoder"
	"github.com/foojank/foojank/internal/services/vessel/scheduler"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	Name       string
	Version    string
	Metadata   map[string]string
	Connection *nats.Conn
	Stream     jetstream.JetStream
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
	rpcSubject := nats.NewInbox()

	connectorOutCh := make(chan connector.Message)
	decoderOutCh := make(chan decoder.Message)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return connector.New(connector.Arguments{
			Name:       s.args.Name,
			Version:    s.args.Version,
			Metadata:   s.args.Metadata,
			RpcSubject: rpcSubject,
			Connection: s.args.Connection,
			OutputCh:   connectorOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return decoder.New(decoder.Arguments{
			InputCh:  connectorOutCh,
			OutputCh: decoderOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return scheduler.New(scheduler.Arguments{
			Connection: s.args.Connection,
			Stream:     s.args.Stream,
			InputCh:    decoderOutCh,
		}).Start(groupCtx)
	})

	return group.Wait()
}
