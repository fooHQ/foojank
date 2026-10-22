package daemon

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/clients/server"
	protoagent "github.com/foohq/foojank/proto/agent"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

const EventStream = "event-stream"

type Client struct {
	srv    *server.Client
	stream string
}

func New(srv *server.Client) *Client {
	return &Client{
		srv:    srv,
		stream: EventStream,
	}
}

func (c *Client) GetAgentHost(ctx context.Context, key string) (AgentHostDirectoryEntry, error) {
	dir, err := c.openAgentHostDirectory(ctx)
	if err != nil {
		return AgentHostDirectoryEntry{}, translate(err)
	}

	host, err := dir.Get(ctx, key)
	if err != nil {
		return AgentHostDirectoryEntry{}, translate(err)
	}

	return host, nil
}

func (c *Client) GetJob(ctx context.Context, key string) (JobDirectoryEntry, error) {
	dir, err := c.openJobDirectory(ctx)
	if err != nil {
		return JobDirectoryEntry{}, translate(err)
	}

	job, err := dir.Get(ctx, key)
	if err != nil {
		return JobDirectoryEntry{}, translate(err)
	}

	return job, nil
}

func (c *Client) ListJobOutputLines(ctx context.Context, job JobDirectoryEntry, count, skip int) ([]byte, error) {
	msgCh := make(chan Message)
	err := c.listStreamMessages(
		ctx,
		[]string{
			protoagent.EvtWorkerStdoutSubject(job.GatewayID, job.AgentID, "*"),
		},
		1+uint64(skip),
		count,
		msgCh,
	)
	if err != nil {
		return nil, translate(err)
	}

	var buff bytes.Buffer
	for msg := range msgCh {
		buff.Write(msg.Data())
	}

	return buff.Bytes(), nil
}

func (c *Client) RequestCreateUser(ctx context.Context, req protodaemon.CreateUserRequest) (protodaemon.CreateUserResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.CreateUserResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.CreateUserSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.CreateUserResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestGetUser(ctx context.Context, req protodaemon.GetUserRequest) (protodaemon.GetUserResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.GetUserResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.GetUserSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.GetUserResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestListUsers(ctx context.Context, req protodaemon.ListUsersRequest) (protodaemon.ListUsersResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.ListUsersResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.ListUsersSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.ListUsersResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestUpdateUser(ctx context.Context, req protodaemon.UpdateUserRequest) (protodaemon.UpdateUserResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.UpdateUserResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.UpdateUserSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.UpdateUserResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestDeleteUser(ctx context.Context, req protodaemon.DeleteUserRequest) (protodaemon.DeleteUserResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.DeleteUserResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.DeleteUserSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.DeleteUserResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestCreateAgent(ctx context.Context, req protodaemon.CreateAgentRequest) (protodaemon.CreateAgentResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.CreateAgentResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.CreateAgentSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.CreateAgentResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestGetAgent(ctx context.Context, req protodaemon.GetAgentRequest) (protodaemon.GetAgentResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.GetAgentResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.GetAgentSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.GetAgentResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestListAgents(ctx context.Context, req protodaemon.ListAgentsRequest) (protodaemon.ListAgentsResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.ListAgentsResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.ListAgentsSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.ListAgentsResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestDeleteAgent(ctx context.Context, req protodaemon.DeleteAgentRequest) (protodaemon.DeleteAgentResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.DeleteAgentResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.DeleteAgentSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.DeleteAgentResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestCreateGateway(ctx context.Context, req protodaemon.CreateGatewayRequest) (protodaemon.CreateGatewayResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.CreateGatewayResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.CreateGatewaySubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.CreateGatewayResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestGetGateway(ctx context.Context, req protodaemon.GetGatewayRequest) (protodaemon.GetGatewayResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.GetGatewayResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.GetGatewaySubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.GetGatewayResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestListGateways(ctx context.Context, req protodaemon.ListGatewaysRequest) (protodaemon.ListGatewaysResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.ListGatewaysResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.ListGatewaysSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.ListGatewaysResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestDeleteGateway(ctx context.Context, req protodaemon.DeleteGatewayRequest) (protodaemon.DeleteGatewayResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.DeleteGatewayResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.DeleteGatewaySubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.DeleteGatewayResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestCreateRole(ctx context.Context, req protodaemon.CreateRoleRequest) (protodaemon.CreateRoleResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.CreateRoleResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.CreateRoleSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.CreateRoleResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestGetRole(ctx context.Context, req protodaemon.GetRoleRequest) (protodaemon.GetRoleResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.GetRoleResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.GetRoleSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.GetRoleResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestListRoles(ctx context.Context, req protodaemon.ListRolesRequest) (protodaemon.ListRolesResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.ListRolesResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.ListRolesSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.ListRolesResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestDeleteRole(ctx context.Context, req protodaemon.DeleteRoleRequest) (protodaemon.DeleteRoleResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.DeleteRoleResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.DeleteRoleSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.DeleteRoleResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestCreateJob(ctx context.Context, req protodaemon.CreateJobRequest) (protodaemon.CreateJobResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.CreateJobResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.CreateJobSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.CreateJobResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestGetJob(ctx context.Context, req protodaemon.GetJobRequest) (protodaemon.GetJobResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.GetJobResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.GetJobSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.GetJobResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestListJobs(ctx context.Context, req protodaemon.ListJobsRequest) (protodaemon.ListJobsResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.ListJobsResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.ListJobsSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.ListJobsResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestCancelJob(ctx context.Context, req protodaemon.CancelJobRequest) (protodaemon.CancelJobResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.CancelJobResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.CancelJobSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.CancelJobResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

func (c *Client) RequestJobOutput(ctx context.Context, req protodaemon.GetJobOutputRequest) (protodaemon.GetJobOutputResponse, error) {
	b, err := protodaemon.Marshal(req)
	if err != nil {
		return protodaemon.GetJobOutputResponse{}, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protodaemon.GetJobOutputSubject(),
		Data:    b,
	})
	if err != nil {
		return protodaemon.GetJobOutputResponse{}, translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
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

const agentHostDirectoryName = "agent-hosts"

func (c *Client) openAgentHostDirectory(ctx context.Context) (*AgentHostDirectory, error) {
	dir, err := c.srv.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: agentHostDirectoryName,
	})
	if err != nil && !errors.Is(err, jetstream.ErrBucketExists) {
		return nil, err
	}
	return &AgentHostDirectory{
		Directory: Directory{
			store: dir,
		},
	}, nil
}

const jobsDirectoryName = "jobs"

func (c *Client) openJobDirectory(ctx context.Context) (*JobDirectory, error) {
	dir, err := c.srv.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: jobsDirectoryName,
	})
	if err != nil && !errors.Is(err, jetstream.ErrBucketExists) {
		return nil, err
	}
	return &JobDirectory{
		Directory: Directory{
			store: dir,
		},
	}, nil
}

func (c *Client) request(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {
	if msg.Header == nil {
		msg.Header = make(nats.Header)
	}
	msg.Header.Set(nats.MsgIdHdr, nuid.Next())
	return c.srv.Conn().RequestMsgWithContext(ctx, msg)
}

func (c *Client) listStreamMessages(ctx context.Context, subjects []string, startSeq uint64, limit int, outputCh chan<- Message) error {
	consumer, err := c.srv.CreateConsumer(ctx, c.stream, jetstream.ConsumerConfig{
		DeliverPolicy:  jetstream.DeliverByStartSequencePolicy,
		AckPolicy:      jetstream.AckNonePolicy,
		MaxAckPending:  1,
		FilterSubjects: subjects,
		OptStartSeq:    startSeq,
	})
	if err != nil {
		return err
	}

	if limit <= 0 {
		limit = 50
	}

	iter, err := consumer.FetchNoWait(limit)
	if err != nil {
		return err
	}

	for msg := range iter.Messages() {
		meta, err := msg.Metadata()
		if err != nil {
			return err
		}

		outputCh <- Message{
			msg:  msg,
			meta: meta,
		}
	}

	close(outputCh)

	err = iter.Error()
	if err != nil {
		return iter.Error()
	}

	return nil
}
