package connector

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"strings"
)

type Arguments struct {
	Name       string
	Version    string
	Metadata   map[string]string
	Connection *nats.Conn
	OutputCh   chan<- Message
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

	rpcSubject := strings.Join([]string{micro.APIPrefix, "RPC", s.args.Name, ms.Info().ID}, ".")
	err = ms.AddEndpoint("rpc", micro.ContextHandler(ctx, s.handler), micro.WithEndpointSubject(rpcSubject))
	if err != nil {
		return err
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
