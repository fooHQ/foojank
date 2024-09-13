package vessel

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Client struct {
	nc *nats.Conn
}

func New(nc *nats.Conn) *Client {
	return &Client{
		nc: nc,
	}
}

type ID struct {
	serviceName string
	serviceID   string
}

func (i ID) infoSubject() string {
	subject, _ := micro.ControlSubject(micro.InfoVerb, i.serviceName, i.serviceID)
	return subject
}

func NewID(serviceName, id string) ID {
	return ID{
		serviceName: serviceName,
		serviceID:   id,
	}
}

type Endpoint struct {
	name    string
	subject string
}

type Service struct {
	ID        ID
	Metadata  map[string]string
	Endpoints map[string]Endpoint
}

func (c *Client) Discover(ctx context.Context, serviceName string) ([]Service, error) {
	subject, err := micro.ControlSubject(micro.InfoVerb, serviceName, "")
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

	var results []Service
loop:
	for {
		select {
		case msg := <-msgCh:
			res, err := c.parseInfo(msg.Data)
			if err != nil {
				continue
			}

			results = append(results, res)

		case <-ctx.Done():
			break loop
		}
	}

	return results, nil
}

func (c *Client) GetInfo(ctx context.Context, id ID) (Service, error) {
	subject := id.infoSubject()
	msg := &nats.Msg{
		Subject: subject,
		Reply:   nats.NewInbox(),
	}
	resp, err := c.nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return Service{}, err
	}

	return c.parseInfo(resp.Data)
}

func (c *Client) CreateWorker(ctx context.Context, s Service) (uint64, error) {
	b, err := NewCreateWorkerRequest()
	if err != nil {
		return 0, err
	}

	rpcEndpoint, ok := s.Endpoints["rpc"]
	if !ok {
		return 0, errors.New("rpc endpoint not found")
	}

	msg := &nats.Msg{
		Subject: rpcEndpoint.subject,
		Reply:   nats.NewInbox(),
		Data:    b,
	}
	resp, err := c.nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return 0, err
	}

	workerID, err := ParseCreateWorkerResponse(resp.Data)
	if err != nil {
		return 0, err
	}

	return workerID, nil
}

func (c *Client) GetWorker(ctx context.Context, s Service, workerID uint64) (ID, error) {
	b, err := NewGetWorkerRequest(workerID)
	if err != nil {
		return ID{}, err
	}

	rpcEndpoint, ok := s.Endpoints["rpc"]
	if !ok {
		return ID{}, errors.New("rpc endpoint not found")
	}

	msg := &nats.Msg{
		Subject: rpcEndpoint.subject,
		Reply:   nats.NewInbox(),
		Data:    b,
	}
	resp, err := c.nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return ID{}, err
	}

	serviceName, serviceID, err := ParseGetWorkerResponse(resp.Data)
	if err != nil {
		return ID{}, err
	}

	return ID{
		serviceName: serviceName,
		serviceID:   serviceID,
	}, nil
}

func (c *Client) parseInfo(b []byte) (Service, error) {
	var data micro.Info
	err := json.Unmarshal(b, &data)
	if err != nil {
		return Service{}, err
	}

	metadata := make(map[string]string, len(data.Metadata))
	for k, v := range data.Metadata {
		metadata[k] = v
	}

	endpoints := make(map[string]Endpoint, len(data.Endpoints))
	for _, endpoint := range data.Endpoints {
		endpoints[endpoint.Name] = Endpoint{
			name:    endpoint.Name,
			subject: endpoint.Subject,
		}
	}

	return Service{
		ID: ID{
			serviceName: data.Name,
			serviceID:   data.ID,
		},
		Metadata:  metadata,
		Endpoints: endpoints,
	}, nil
}
