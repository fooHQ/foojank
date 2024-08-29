package vessel

import (
	"context"
	"github.com/foojank/foojank/internal/services/vessel/connector"
	"github.com/foojank/foojank/internal/services/vessel/decoder"
	"github.com/foojank/foojank/internal/services/vessel/scheduler"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	Connection *nats.Conn
	Connector  connector.Arguments
	Decoder    decoder.Arguments
	Scheduler  scheduler.Arguments
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
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return connector.New(s.args.Connector).Start(groupCtx)
	})
	group.Go(func() error {
		return decoder.New(s.args.Decoder).Start(groupCtx)
	})
	group.Go(func() error {
		return scheduler.New(s.args.Scheduler).Start(groupCtx)
	})

	return group.Wait()
}
