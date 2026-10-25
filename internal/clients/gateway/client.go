package gateway

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/clients/server"
	protogw "github.com/foohq/foojank/proto/gateway"
)

type Client struct {
	srv *server.Client
}

func New(srv *server.Client) *Client {
	return &Client{
		srv: srv,
	}
}

func (c *Client) RegisterAgent(ctx context.Context, gatewayID string, req protogw.RegisterAgentRequest) (protogw.RegisterAgentResponse, error) {
	data, err := c.request(ctx, protogw.RegisterAgentSubject(gatewayID), req)
	if err != nil {
		return protogw.RegisterAgentResponse{}, err
	}

	v, ok := data.(protogw.RegisterAgentResponse)
	if !ok {
		return protogw.RegisterAgentResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protogw.RegisterAgentResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) UnregisterAgent(ctx context.Context, gatewayID string, req protogw.UnregisterAgentRequest) (protogw.UnregisterAgentResponse, error) {
	data, err := c.request(ctx, protogw.UnregisterAgentSubject(gatewayID), req)
	if err != nil {
		return protogw.UnregisterAgentResponse{}, err
	}

	v, ok := data.(protogw.UnregisterAgentResponse)
	if !ok {
		return protogw.UnregisterAgentResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protogw.UnregisterAgentResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) CreateJob(ctx context.Context, gatewayID string, req protogw.CreateJobRequest) (protogw.CreateJobResponse, error) {
	data, err := c.request(ctx, protogw.CreateJobSubject(gatewayID), req)
	if err != nil {
		return protogw.CreateJobResponse{}, err
	}

	v, ok := data.(protogw.CreateJobResponse)
	if !ok {
		return protogw.CreateJobResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protogw.CreateJobResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) CancelJob(ctx context.Context, gatewayID string, req protogw.CancelJobRequest) (protogw.CancelJobResponse, error) {
	data, err := c.request(ctx, protogw.CancelJobSubject(gatewayID), req)
	if err != nil {
		return protogw.CancelJobResponse{}, err
	}

	v, ok := data.(protogw.CancelJobResponse)
	if !ok {
		return protogw.CancelJobResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protogw.CancelJobResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) WriteJobInput(ctx context.Context, gatewayID string, req protogw.JobInputRequest) (protogw.JobInputResponse, error) {
	data, err := c.request(ctx, protogw.JobInputSubject(gatewayID), req)
	if err != nil {
		return protogw.JobInputResponse{}, err
	}

	v, ok := data.(protogw.JobInputResponse)
	if !ok {
		return protogw.JobInputResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protogw.JobInputResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) ReadJobOutput(ctx context.Context, gatewayID string, req protogw.JobOutputRequest) (protogw.JobOutputResponse, error) {
	data, err := c.request(ctx, protogw.JobOutputSubject(gatewayID), req)
	if err != nil {
		return protogw.JobOutputResponse{}, err
	}

	v, ok := data.(protogw.JobOutputResponse)
	if !ok {
		return protogw.JobOutputResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protogw.JobOutputResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) RegisterGateway(ctx context.Context, gatewayID string) error {
	_, err := c.srv.CreateConsumer(ctx, gatewayID, jetstream.ConsumerConfig{
		Durable:       gatewayID,
		DeliverPolicy: jetstream.DeliverLastPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
		FilterSubjects: []string{
			protogw.RegisterAgentSubject(gatewayID),
			protogw.UnregisterAgentSubject(gatewayID),
		},
	})
	return err
}

func (c *Client) UnregisterGateway(ctx context.Context, gatewayID string) error {
	return c.srv.DeleteConsumer(ctx, gatewayID, gatewayID)
}

func (c *Client) request(ctx context.Context, subject string, req any) (any, error) {
	b, err := protogw.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.srv.Conn().RequestMsgWithContext(ctx, &nats.Msg{
		Subject: subject,
		Header: nats.Header{
			nats.MsgIdHdr: []string{nuid.Next()},
		},
		Data: b,
	})
	if err != nil {
		return nil, err
	}

	data, err := protogw.Unmarshal(resp.Data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
