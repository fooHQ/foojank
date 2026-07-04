package daemon

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	protoagent "github.com/foohq/foojank-proto/go/agent"
	protogw "github.com/foohq/foojank-proto/go/gateway"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/clients/server"
)

const eventStream = "event-stream"

type Client struct {
	srv    *server.Client
	stream string
}

func New(srv *server.Client) *Client {
	return &Client{
		srv:    srv,
		stream: eventStream,
	}
}

func (c *Client) InitDaemon(ctx context.Context) (err error) {
	streamName := c.stream
	_, err = c.srv.CreateStream(ctx, jetstream.StreamConfig{
		Name: streamName,
		Subjects: []string{
			protoagent.CmdStartWorkerSubject("*", "*", "*"),
			protoagent.CmdStopWorkerSubject("*", "*", "*"),
			protoagent.CmdWriteStdinSubject("*", "*", "*"),
			protoagent.EvtStartWorkerSubject("*", "*", "*"),
			protoagent.EvtStopWorkerSubject("*", "*", "*"),
			protoagent.EvtWorkerStdoutSubject("*", "*", "*"),
			protoagent.EvtWorkerStatusSubject("*", "*", "*"),
			protoagent.EvtAgentInfoSubject("*", "*"),
		},
	})
	if err != nil {
		return translate(err)
	}
	defer func() {
		if err == nil {
			return
		}
		_ = c.srv.DeleteStream(ctx, c.stream)
	}()

	consumerName := c.srv.UserID()
	_, err = c.srv.CreateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: jetstream.DeliverLastPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
	})
	if err != nil {
		return translate(err)
	}
	defer func() {
		if err == nil {
			return
		}
		_ = c.srv.DeleteConsumer(ctx, streamName, consumerName)
	}()

	return nil
}

func (c *Client) CreateAgent(ctx context.Context, agent AgentDirectoryEntry) error {
	dir, err := c.openAgentDirectory(ctx)
	if err != nil {
		return translate(err)
	}

	err = dir.Create(ctx, agent)
	if err != nil {
		return translate(err)
	}

	return nil
}

func (c *Client) RemoveAgent(ctx context.Context, agent AgentDirectoryEntry) error {
	dir, err := c.openAgentDirectory(ctx)
	if err != nil {
		return translate(err)
	}

	err = dir.Delete(ctx, agent.ID)
	if err != nil && !errors.Is(err, jetstream.ErrKeyNotFound) {
		return translate(err)
	}

	return nil
}

func (c *Client) GetAgent(ctx context.Context, key string) (AgentDirectoryEntry, error) {
	dir, err := c.openAgentDirectory(ctx)
	if err != nil {
		return AgentDirectoryEntry{}, translate(err)
	}

	agent, err := dir.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return AgentDirectoryEntry{}, ErrAgentNotFound
		}
		return AgentDirectoryEntry{}, translate(err)
	}

	return agent, nil
}

func (c *Client) ListAgents(ctx context.Context) ([]AgentDirectoryEntry, error) {
	dir, err := c.openAgentDirectory(ctx)
	if err != nil {
		return nil, translate(err)
	}

	entries, err := dir.List(ctx)
	if err != nil {
		return nil, translate(err)
	}

	return entries, nil
}

func (c *Client) ListAgentHosts(ctx context.Context) ([]AgentHostDirectoryEntry, error) {
	dir, err := c.openAgentHostDirectory(ctx)
	if err != nil {
		return nil, translate(err)
	}

	entries, err := dir.List(ctx)
	if err != nil {
		return nil, translate(err)
	}

	return entries, nil
}

func (c *Client) CreateGateway(ctx context.Context, gateway GatewayDirectoryEntry) error {
	dir, err := c.openGatewayDirectory(ctx)
	if err != nil {
		return translate(err)
	}

	err = dir.Create(ctx, gateway)
	if err != nil {
		return translate(err)
	}

	return nil
}

func (c *Client) RemoveGateway(ctx context.Context, gateway GatewayDirectoryEntry) error {
	dir, err := c.openGatewayDirectory(ctx)
	if err != nil {
		return translate(err)
	}

	err = dir.Delete(ctx, gateway.ID)
	if err != nil && !errors.Is(err, jetstream.ErrKeyNotFound) {
		return translate(err)
	}

	return nil
}

func (c *Client) GetGateway(ctx context.Context, key string) (GatewayDirectoryEntry, error) {
	dir, err := c.openGatewayDirectory(ctx)
	if err != nil {
		return GatewayDirectoryEntry{}, translate(err)
	}

	gateway, err := dir.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return GatewayDirectoryEntry{}, ErrGatewayNotFound
		}
		return GatewayDirectoryEntry{}, translate(err)
	}

	return gateway, nil
}

func (c *Client) ListGateways(ctx context.Context) ([]GatewayDirectoryEntry, error) {
	dir, err := c.openGatewayDirectory(ctx)
	if err != nil {
		return nil, translate(err)
	}

	entries, err := dir.List(ctx)
	if err != nil {
		return nil, translate(err)
	}

	return entries, nil
}

func (c *Client) CreateJob(ctx context.Context, job JobDirectoryEntry) error {
	dir, err := c.openJobDirectory(ctx)
	if err != nil {
		return translate(err)
	}

	err = dir.Create(ctx, job)
	if err != nil {
		return translate(err)
	}

	return nil
}

func (c *Client) GetJob(ctx context.Context, key string) (JobDirectoryEntry, error) {
	dir, err := c.openJobDirectory(ctx)
	if err != nil {
		return JobDirectoryEntry{}, translate(err)
	}

	job, err := dir.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return JobDirectoryEntry{}, ErrJobNotFound
		}
		return JobDirectoryEntry{}, translate(err)
	}

	return job, nil
}

func (c *Client) ListJobsByAgentID(ctx context.Context, agentID string) ([]JobDirectoryEntry, error) {
	dir, err := c.openJobDirectory(ctx)
	if err != nil {
		return nil, translate(err)
	}

	entries, err := dir.ListByAgentID(ctx, agentID)
	if err != nil {
		return nil, translate(err)
	}

	return entries, nil
}

func (c *Client) ListJobs(ctx context.Context) ([]JobDirectoryEntry, error) {
	dir, err := c.openJobDirectory(ctx)
	if err != nil {
		return nil, translate(err)
	}

	entries, err := dir.List(ctx)
	if err != nil {
		return nil, translate(err)
	}

	return entries, nil
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

func (c *Client) RegisterGateway(ctx context.Context, gateway GatewayDirectoryEntry) error {
	consumerName := gateway.ID
	_, err := c.srv.CreateConsumer(ctx, c.stream, jetstream.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: jetstream.DeliverLastPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
		FilterSubjects: []string{
			protoagent.CmdStartWorkerSubject(gateway.ID, "*", "*"),
			protoagent.CmdStopWorkerSubject(gateway.ID, "*", "*"),
			protoagent.CmdWriteStdinSubject(gateway.ID, "*", "*"),
			protogw.CmdRegisterAgentSubject(gateway.ID, "*"),
			protogw.CmdUnregisterAgentSubject(gateway.ID, "*"),
		},
	})
	if err != nil {
		return translate(err)
	}

	return nil
}

func (c *Client) UnregisterGateway(ctx context.Context, gateway GatewayDirectoryEntry) error {
	// TODO: verify that current users of the API are able to handle the error caused by non-existent consumer!
	consumerName := gateway.ID
	err := c.srv.DeleteConsumer(ctx, c.stream, consumerName)
	if err != nil {
		return translate(err)
	}
	return nil
}

func (c *Client) RequestRegisterAgent(ctx context.Context, agent AgentDirectoryEntry) (map[string]string, error) {
	b, err := protogw.Marshal(protogw.Envelope{
		Subject: protogw.CmdRegisterAgentSubject(agent.GatewayID, agent.ID),
		Payload: protogw.RegisterAgentRequest{
			// TODO: add agentID etc
			Properties: propertiesToList(agent.Config.Extra),
		},
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protogw.CmdRegisterAgentSubject(agent.GatewayID, agent.ID),
		Data:    b,
	})
	if err != nil {
		return nil, translate(err)
	}

	data, err := protogw.Unmarshal(resp.Data)
	if err != nil {
		return nil, err
	}

	v, ok := data.Payload.(protogw.RegisterAgentResponse)
	if !ok {
		return nil, fmt.Errorf("invalid payload: %T", data.Payload)
	}

	if v.Error != nil {
		return nil, v.Error
	}

	return propertiesToMap(v.Properties), nil
}

func (c *Client) RequestUnregisterAgent(ctx context.Context, agent AgentDirectoryEntry) error {
	b, err := protogw.Marshal(protogw.Envelope{
		Subject: protogw.CmdUnregisterAgentSubject(agent.GatewayID, agent.ID),
		Payload: protogw.UnregisterAgentRequest{
			Properties: propertiesToList(agent.Config.Extra),
		},
	})
	if err != nil {
		return err
	}

	resp, err := c.request(ctx, &nats.Msg{
		Subject: protogw.CmdUnregisterAgentSubject(agent.GatewayID, agent.ID),
		Data:    b,
	})
	if err != nil {
		return translate(err)
	}

	data, err := protogw.Unmarshal(resp.Data)
	if err != nil {
		return err
	}

	v, ok := data.Payload.(protogw.UnregisterAgentResponse)
	if !ok {
		return fmt.Errorf("invalid payload: %T", data.Payload)
	}

	if v.Error != nil {
		return v.Error
	}

	return nil
}

func (c *Client) PublishStartWorkerRequest(ctx context.Context, job JobDirectoryEntry) error {
	b, err := protoagent.Marshal(protoagent.Envelope{
		Subject: "",
		Payload: protoagent.StartWorkerRequest{
			Command: job.Config.Command,
			Args:    job.Config.Args,
			Env:     job.Config.Env,
		},
	})
	if err != nil {
		return err
	}

	err = c.publishMsgToStream(
		ctx,
		&nats.Msg{
			Subject: protoagent.CmdStartWorkerSubject(job.GatewayID, job.ID, job.WorkerID),
			Data:    b,
		},
	)
	if err != nil {
		return translate(err)
	}

	return nil
}

func (c *Client) PublishStopWorkerRequest(ctx context.Context, job JobDirectoryEntry) error {
	b, err := protoagent.Marshal(protoagent.Envelope{
		Subject: "",
		Payload: protoagent.StopWorkerRequest{},
	})
	if err != nil {
		return err
	}

	err = c.publishMsgToStream(
		ctx,
		&nats.Msg{
			Subject: protoagent.CmdStopWorkerSubject(job.GatewayID, job.AgentID, job.WorkerID),
			Data:    b,
		},
	)
	if err != nil {
		return translate(err)
	}

	return nil
}

func (c *Client) PublishWorkerStdin(ctx context.Context, job JobDirectoryEntry) error {
	b, err := protoagent.Marshal(protoagent.Envelope{
		Subject: "",
		Payload: protoagent.UpdateWorkerStdio{},
	})
	if err != nil {
		return err
	}

	err = c.publishMsgToStream(
		ctx,
		&nats.Msg{
			Subject: protoagent.CmdWriteStdinSubject(job.GatewayID, job.AgentID, job.WorkerID),
			Data:    b,
		},
	)
	if err != nil {
		return translate(err)
	}

	return nil
}

func (c *Client) publishMsgToStream(ctx context.Context, msg *nats.Msg) error {
	if msg.Header == nil {
		msg.Header = make(nats.Header)
	}
	msg.Header.Set(nats.MsgIdHdr, nuid.Next())
	_, err := c.srv.PublishMsg(
		ctx,
		msg,
		jetstream.WithExpectStream(c.stream),
	)
	if err != nil {
		return err
	}
	return nil
}

const agentDirectoryName = "agents"

func (c *Client) openAgentDirectory(ctx context.Context) (*AgentDirectory, error) {
	dir, err := c.srv.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: agentDirectoryName,
	})
	if err != nil && !errors.Is(err, jetstream.ErrBucketExists) {
		return nil, err
	}
	return &AgentDirectory{
		Directory: Directory{
			store: dir,
		},
	}, nil
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

const gatewayDirectoryName = "gateways"

func (c *Client) openGatewayDirectory(ctx context.Context) (*GatewayDirectory, error) {
	dir, err := c.srv.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: gatewayDirectoryName,
	})
	if err != nil && !errors.Is(err, jetstream.ErrBucketExists) {
		return nil, err
	}
	return &GatewayDirectory{
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

func propertiesToList(props map[string]string) []protogw.Property {
	var list []protogw.Property
	for k, v := range props {
		list = append(list, protogw.Property{Key: k, Value: v})
	}
	return list
}

func propertiesToMap(props []protogw.Property) map[string]string {
	m := make(map[string]string, len(props))
	for _, p := range props {
		m[p.Key] = p.Value
	}
	return m
}

func InboxName(name string) string {
	return "_INBOX_" + name
}

func NewGatewayPermissions(gatewayID string) jwt.Permissions {
	return jwt.Permissions{
		Pub: jwt.Permission{
			Allow: []string{
				fmt.Sprintf("$JS.ACK.%s.>", gatewayID),
				fmt.Sprintf("$JS.API.STREAM.INFO.%s", gatewayID),
				fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", gatewayID),
				fmt.Sprintf("$JS.API.STREAM.PURGE.OBJ_%s", gatewayID),
				fmt.Sprintf("$JS.API.CONSUMER.INFO.%s.*", gatewayID),
				fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.%s.*", gatewayID),
				fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.M.*", gatewayID, gatewayID),
				fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.>", gatewayID),
				fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", gatewayID),
				fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.>", gatewayID),
				fmt.Sprintf("$O.%s.M.*", gatewayID),
				fmt.Sprintf("$O.%s.C.*", gatewayID),
			},
		},
		Sub: jwt.Permission{
			Allow: []string{
				fmt.Sprintf("%s.>", InboxName(gatewayID)),
			},
		},
	}
}
