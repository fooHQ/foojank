package adapter

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "FOOJANK_GATEWAY_ADAPTER",
	MagicCookieValue: "f00j4nk",
}

var _ plugin.Plugin = (*Plugin)(nil)

type Plugin struct {
	impl Adapter
}

func NewPlugin(adapter Adapter) *Plugin {
	return &Plugin{
		impl: adapter,
	}
}

func (p *Plugin) Server(b *plugin.MuxBroker) (any, error) {
	return &RPCServer{
		impl:   p.impl,
		broker: b,
	}, nil
}

func (p *Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &RPCClient{
		client: c,
		broker: b,
	}, nil
}
