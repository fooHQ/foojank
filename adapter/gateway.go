package adapter

import "net/rpc"

type Gateway interface {
	GetMessages() ([]string, error)
}

type GatewayRPCClient struct {
	cli *rpc.Client
}

func (c *GatewayRPCClient) GetMessages() ([]string, error) {
	var resp []string
	err := c.cli.Call("Plugin.GetMessages", new(any), &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type GatewayRPCServer struct {
	impl Gateway
}

func (s *GatewayRPCServer) GetMessages(_ any, resp *[]string) error {
	v, err := s.impl.GetMessages()
	if err != nil {
		return err
	}
	*resp = v
	return nil
}
