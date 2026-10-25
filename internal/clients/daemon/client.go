package daemon

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/clients/server"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type Client struct {
	srv *server.Client
}

func New(srv *server.Client) *Client {
	return &Client{
		srv: srv,
	}
}

func (c *Client) CreateUser(ctx context.Context, req protodaemon.CreateUserRequest) (protodaemon.CreateUserResponse, error) {
	data, err := c.request(ctx, protodaemon.CreateUserSubject(), req)
	if err != nil {
		return protodaemon.CreateUserResponse{}, err
	}

	v, ok := data.(protodaemon.CreateUserResponse)
	if !ok {
		return protodaemon.CreateUserResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.CreateUserResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) GetUser(ctx context.Context, req protodaemon.GetUserRequest) (protodaemon.GetUserResponse, error) {
	data, err := c.request(ctx, protodaemon.GetUserSubject(), req)
	if err != nil {
		return protodaemon.GetUserResponse{}, err
	}

	v, ok := data.(protodaemon.GetUserResponse)
	if !ok {
		return protodaemon.GetUserResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.GetUserResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) ListUsers(ctx context.Context, req protodaemon.ListUsersRequest) (protodaemon.ListUsersResponse, error) {
	data, err := c.request(ctx, protodaemon.ListUsersSubject(), req)
	if err != nil {
		return protodaemon.ListUsersResponse{}, err
	}

	v, ok := data.(protodaemon.ListUsersResponse)
	if !ok {
		return protodaemon.ListUsersResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.ListUsersResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) UpdateUser(ctx context.Context, req protodaemon.UpdateUserRequest) (protodaemon.UpdateUserResponse, error) {
	data, err := c.request(ctx, protodaemon.UpdateUserSubject(), req)
	if err != nil {
		return protodaemon.UpdateUserResponse{}, err
	}

	v, ok := data.(protodaemon.UpdateUserResponse)
	if !ok {
		return protodaemon.UpdateUserResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.UpdateUserResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) DeleteUser(ctx context.Context, req protodaemon.DeleteUserRequest) (protodaemon.DeleteUserResponse, error) {
	data, err := c.request(ctx, protodaemon.DeleteUserSubject(), req)
	if err != nil {
		return protodaemon.DeleteUserResponse{}, err
	}

	v, ok := data.(protodaemon.DeleteUserResponse)
	if !ok {
		return protodaemon.DeleteUserResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.DeleteUserResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) CreateAgent(ctx context.Context, req protodaemon.CreateAgentRequest) (protodaemon.CreateAgentResponse, error) {
	data, err := c.request(ctx, protodaemon.CreateAgentSubject(), req)
	if err != nil {
		return protodaemon.CreateAgentResponse{}, err
	}

	v, ok := data.(protodaemon.CreateAgentResponse)
	if !ok {
		return protodaemon.CreateAgentResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.CreateAgentResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) GetAgent(ctx context.Context, req protodaemon.GetAgentRequest) (protodaemon.GetAgentResponse, error) {
	data, err := c.request(ctx, protodaemon.GetAgentSubject(), req)
	if err != nil {
		return protodaemon.GetAgentResponse{}, err
	}

	v, ok := data.(protodaemon.GetAgentResponse)
	if !ok {
		return protodaemon.GetAgentResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.GetAgentResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) ListAgents(ctx context.Context, req protodaemon.ListAgentsRequest) (protodaemon.ListAgentsResponse, error) {
	data, err := c.request(ctx, protodaemon.ListAgentsSubject(), req)
	if err != nil {
		return protodaemon.ListAgentsResponse{}, err
	}

	v, ok := data.(protodaemon.ListAgentsResponse)
	if !ok {
		return protodaemon.ListAgentsResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.ListAgentsResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) DeleteAgent(ctx context.Context, req protodaemon.DeleteAgentRequest) (protodaemon.DeleteAgentResponse, error) {
	data, err := c.request(ctx, protodaemon.DeleteAgentSubject(), req)
	if err != nil {
		return protodaemon.DeleteAgentResponse{}, err
	}

	v, ok := data.(protodaemon.DeleteAgentResponse)
	if !ok {
		return protodaemon.DeleteAgentResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.DeleteAgentResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) CreateGateway(ctx context.Context, req protodaemon.CreateGatewayRequest) (protodaemon.CreateGatewayResponse, error) {
	data, err := c.request(ctx, protodaemon.CreateGatewaySubject(), req)
	if err != nil {
		return protodaemon.CreateGatewayResponse{}, err
	}

	v, ok := data.(protodaemon.CreateGatewayResponse)
	if !ok {
		return protodaemon.CreateGatewayResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.CreateGatewayResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) GetGateway(ctx context.Context, req protodaemon.GetGatewayRequest) (protodaemon.GetGatewayResponse, error) {
	data, err := c.request(ctx, protodaemon.GetGatewaySubject(), req)
	if err != nil {
		return protodaemon.GetGatewayResponse{}, err
	}

	v, ok := data.(protodaemon.GetGatewayResponse)
	if !ok {
		return protodaemon.GetGatewayResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.GetGatewayResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) ListGateways(ctx context.Context, req protodaemon.ListGatewaysRequest) (protodaemon.ListGatewaysResponse, error) {
	data, err := c.request(ctx, protodaemon.ListGatewaysSubject(), req)
	if err != nil {
		return protodaemon.ListGatewaysResponse{}, err
	}

	v, ok := data.(protodaemon.ListGatewaysResponse)
	if !ok {
		return protodaemon.ListGatewaysResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.ListGatewaysResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) DeleteGateway(ctx context.Context, req protodaemon.DeleteGatewayRequest) (protodaemon.DeleteGatewayResponse, error) {
	data, err := c.request(ctx, protodaemon.DeleteGatewaySubject(), req)
	if err != nil {
		return protodaemon.DeleteGatewayResponse{}, err
	}

	v, ok := data.(protodaemon.DeleteGatewayResponse)
	if !ok {
		return protodaemon.DeleteGatewayResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.DeleteGatewayResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) CreateRole(ctx context.Context, req protodaemon.CreateRoleRequest) (protodaemon.CreateRoleResponse, error) {
	data, err := c.request(ctx, protodaemon.CreateRoleSubject(), req)
	if err != nil {
		return protodaemon.CreateRoleResponse{}, err
	}

	v, ok := data.(protodaemon.CreateRoleResponse)
	if !ok {
		return protodaemon.CreateRoleResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.CreateRoleResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) GetRole(ctx context.Context, req protodaemon.GetRoleRequest) (protodaemon.GetRoleResponse, error) {
	data, err := c.request(ctx, protodaemon.GetRoleSubject(), req)
	if err != nil {
		return protodaemon.GetRoleResponse{}, err
	}

	v, ok := data.(protodaemon.GetRoleResponse)
	if !ok {
		return protodaemon.GetRoleResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.GetRoleResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) ListRoles(ctx context.Context, req protodaemon.ListRolesRequest) (protodaemon.ListRolesResponse, error) {
	data, err := c.request(ctx, protodaemon.ListRolesSubject(), req)
	if err != nil {
		return protodaemon.ListRolesResponse{}, err
	}

	v, ok := data.(protodaemon.ListRolesResponse)
	if !ok {
		return protodaemon.ListRolesResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.ListRolesResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) DeleteRole(ctx context.Context, req protodaemon.DeleteRoleRequest) (protodaemon.DeleteRoleResponse, error) {
	data, err := c.request(ctx, protodaemon.DeleteRoleSubject(), req)
	if err != nil {
		return protodaemon.DeleteRoleResponse{}, err
	}

	v, ok := data.(protodaemon.DeleteRoleResponse)
	if !ok {
		return protodaemon.DeleteRoleResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.DeleteRoleResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) CreateJob(ctx context.Context, req protodaemon.CreateJobRequest) (protodaemon.CreateJobResponse, error) {
	data, err := c.request(ctx, protodaemon.CreateJobSubject(), req)
	if err != nil {
		return protodaemon.CreateJobResponse{}, err
	}

	v, ok := data.(protodaemon.CreateJobResponse)
	if !ok {
		return protodaemon.CreateJobResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.CreateJobResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) GetJob(ctx context.Context, req protodaemon.GetJobRequest) (protodaemon.GetJobResponse, error) {
	data, err := c.request(ctx, protodaemon.GetJobSubject(), req)
	if err != nil {
		return protodaemon.GetJobResponse{}, err
	}

	v, ok := data.(protodaemon.GetJobResponse)
	if !ok {
		return protodaemon.GetJobResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.GetJobResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) ListJobs(ctx context.Context, req protodaemon.ListJobsRequest) (protodaemon.ListJobsResponse, error) {
	data, err := c.request(ctx, protodaemon.ListJobsSubject(), req)
	if err != nil {
		return protodaemon.ListJobsResponse{}, err
	}

	v, ok := data.(protodaemon.ListJobsResponse)
	if !ok {
		return protodaemon.ListJobsResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.ListJobsResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) CancelJob(ctx context.Context, req protodaemon.CancelJobRequest) (protodaemon.CancelJobResponse, error) {
	data, err := c.request(ctx, protodaemon.CancelJobSubject(), req)
	if err != nil {
		return protodaemon.CancelJobResponse{}, err
	}

	v, ok := data.(protodaemon.CancelJobResponse)
	if !ok {
		return protodaemon.CancelJobResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.CancelJobResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) ReadJobOutput(ctx context.Context, req protodaemon.GetJobOutputRequest) (protodaemon.GetJobOutputResponse, error) {
	data, err := c.request(ctx, protodaemon.GetJobOutputSubject(), req)
	if err != nil {
		return protodaemon.GetJobOutputResponse{}, err
	}

	v, ok := data.(protodaemon.GetJobOutputResponse)
	if !ok {
		return protodaemon.GetJobOutputResponse{}, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return protodaemon.GetJobOutputResponse{}, v.Error
	}

	return v, nil
}

func (c *Client) request(ctx context.Context, subject string, req any) (any, error) {
	b, err := protodaemon.Marshal(req)
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
		return nil, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
