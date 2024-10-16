package connector

import (
	"context"
	"github.com/foojank/foojank/internal/log"
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
	InfoCh       chan<- InfoMessage
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
	defer func() {
		err := ms.Stop()
		if err != nil {
			log.Debug(err.Error())
			return
		}
	}()

	err = ms.AddEndpoint("data", micro.ContextHandler(ctx, s.handler), micro.WithEndpointSubject(s.args.DataSubject))
	if err != nil {
		return err
	}

	err = ms.AddEndpoint("stdin", micro.ContextHandler(ctx, s.handler), micro.WithEndpointSubject(s.args.StdinSubject))
	if err != nil {
		return err
	}

	info := InfoMessage{
		serviceName: ms.Info().Name,
		serviceID:   ms.Info().ID,
	}
	select {
	case s.args.InfoCh <- info:
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
