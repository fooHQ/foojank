package vessel

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Client struct {
	nc          *nats.Conn
	serviceName string
}

func New(nc *nats.Conn, serviceName string) *Client {
	return &Client{
		nc:          nc,
		serviceName: serviceName,
	}
}

type AgentInfo struct {
	serviceName string
	ID          string
}

func (c *Client) Discover(ctx context.Context) ([]AgentInfo, error) {
	subject, err := micro.ControlSubject(micro.InfoVerb, c.serviceName, "")
	if err != nil {
		return nil, err
	}

	inbox := nats.NewInbox()
	msgCh := make(chan *nats.Msg, 1024)
	sub, err := c.nc.ChanSubscribe(inbox, msgCh)
	if err != nil {
		return nil, err
	}
	defer sub.Drain()

	reqMsg := &nats.Msg{
		Subject: subject,
		Reply:   inbox,
	}
	err = c.nc.PublishMsg(reqMsg)
	if err != nil {
		return nil, err
	}

	var result []AgentInfo
loop:
	for {
		select {
		case msg := <-msgCh:
			var data micro.Info
			err := json.Unmarshal(msg.Data, &data)
			if err != nil {
				continue
			}

			result = append(result, AgentInfo{
				serviceName: c.serviceName,
				ID:          data.ID,
			})

		case <-ctx.Done():
			break loop
		}
	}

	return result, nil
}
