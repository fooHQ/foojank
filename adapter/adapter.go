package adapter

import (
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"
)

type Adapter interface {
	Setup() error
	AttachGateway(Gateway) error
	RegisterAgent() error
	UnregisterAgent() error
	Shutdown() error
}

type RPCClient struct {
	broker *plugin.MuxBroker
	client *rpc.Client
}

func (p *RPCClient) Setup() error {
	var reply struct{}
	return p.client.Call("Plugin.Setup", new(any), &reply)
}

func (p *RPCClient) AttachGateway(gateway Gateway) error {
	id := p.broker.NextId()

	go p.broker.AcceptAndServe(id, &GatewayRPCServer{
		impl: gateway,
	})

	var reply struct{}
	err := p.client.Call("Plugin.AttachGateway", id, &reply)
	if err != nil {
		return err
	}

	return nil
}

func (p *RPCClient) RegisterAgent() error {
	var reply struct{}
	return p.client.Call("Plugin.RegisterAgent", new(any), &reply)
}

func (p *RPCClient) UnregisterAgent() error {
	var reply struct{}
	return p.client.Call("Plugin.UnregisterAgent", new(any), &reply)
}

func (p *RPCClient) Shutdown() error {
	var reply struct{}
	return p.client.Call("Plugin.Shutdown", new(any), &reply)
}

type RPCServer struct {
	impl   Adapter
	broker *plugin.MuxBroker
	client *rpc.Client
	mu     sync.Mutex
}

func (s *RPCServer) Setup(_ any, _ *struct{}) error {
	return s.impl.Setup()
}

func (s *RPCServer) AttachGateway(id uint32, _ *struct{}) error {
	conn, err := s.broker.Dial(id)
	if err != nil {
		return err
	}

	cli := rpc.NewClient(conn)

	s.mu.Lock()
	oldCli := s.client
	s.client = cli
	s.mu.Unlock()

	if oldCli != nil {
		_ = oldCli.Close()
	}

	return s.impl.AttachGateway(&GatewayRPCClient{
		cli: cli,
	})
}

func (s *RPCServer) RegisterAgent(_ any, _ *struct{}) error {
	return s.impl.RegisterAgent()
}

func (s *RPCServer) UnregisterAgent(_ any, _ *struct{}) error {
	return s.impl.UnregisterAgent()
}

func (s *RPCServer) Shutdown(_ any, _ *struct{}) error {
	err := s.impl.Shutdown()

	s.mu.Lock()
	cli := s.client
	s.mu.Unlock()

	if cli != nil {
		_ = cli.Close()
	}

	return err
}
