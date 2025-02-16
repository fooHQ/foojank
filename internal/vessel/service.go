package vessel

import (
	"context"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/connector"
	"github.com/foohq/foojank/internal/vessel/decoder"
	"github.com/foohq/foojank/internal/vessel/scheduler"
)

type Arguments struct {
	Name       string
	Version    string
	Metadata   map[string]string
	Connection *nats.Conn
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
	connectorOutCh := make(chan connector.Message)
	decoderOutCh := make(chan decoder.Message)
	rpcSubject := s.args.Name + "." + "RPC"

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return connector.New(connector.Arguments{
			Name:       s.args.Name,
			Version:    s.args.Version,
			Metadata:   s.args.Metadata,
			RPCSubject: rpcSubject,
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
			InputCh:    decoderOutCh,
		}).Start(groupCtx)
	})

	return group.Wait()
}
