package connector

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Arguments struct {
	Name         string
	Version      string
	Metadata     map[string]string
	StdinSubject string
	DataSubject  string
	Connection   *nats.Conn
	IDCh         chan<- string
	OutputCh     chan<- Message
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
	ms, err := micro.AddService(s.args.Connection, micro.Config{
		Name:     s.args.Name,
		Version:  s.args.Version,
		Metadata: s.args.Metadata,
	})
	if err != nil {
		return err
	}
	defer ms.Stop()

	err = ms.AddEndpoint("data", micro.ContextHandler(ctx, s.handler), micro.WithEndpointSubject(s.args.DataSubject))
	if err != nil {
		return err
	}

	err = ms.AddEndpoint("stdin", micro.ContextHandler(ctx, s.handler), micro.WithEndpointSubject(s.args.StdinSubject))
	if err != nil {
		return err
	}

	select {
	case s.args.IDCh <- ms.Info().ID:
	case <-ctx.Done():
		return nil
	}

	<-ctx.Done()
	return nil
}

func (s *Service) handler(ctx context.Context, req micro.Request) {
	msg := Message{
		req: req,
	}
	select {
	case s.args.OutputCh <- msg:
	case <-ctx.Done():
		return
	}
}
