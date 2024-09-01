package client

import (
	"context"
	"github.com/foojank/foojank/internal/application"
	"github.com/foojank/foojank/internal/services/client/reader"
	"github.com/foojank/foojank/internal/services/client/router"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
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
	app := application.New(s.args.Connection)
	terminalOutCh := make(chan reader.Message, 1)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return reader.New(reader.Arguments{
			OutputCh: terminalOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return router.New(router.Arguments{
			App:        app,
			Connection: s.args.Connection,
			InputCh:    terminalOutCh,
		}).Start(groupCtx)
	})

	return group.Wait()
}
