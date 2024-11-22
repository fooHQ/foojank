package vessel

import (
	"context"
	connector2 "github.com/foohq/foojank/internal/vessel/connector"
	decoder2 "github.com/foohq/foojank/internal/vessel/decoder"
	"github.com/foohq/foojank/internal/vessel/scheduler"
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
	connectorOutCh := make(chan connector2.Message)
	decoderOutCh := make(chan decoder2.Message)
	rpcSubject := s.args.Name + "." + "RPC"

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return connector2.New(connector2.Arguments{
			Name:       s.args.Name,
			Version:    s.args.Version,
			Metadata:   s.args.Metadata,
			RpcSubject: rpcSubject,
			Connection: s.args.Connection,
			OutputCh:   connectorOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return decoder2.New(decoder2.Arguments{
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
