package agent

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/router"
	"github.com/foohq/foojank/proto"
)

type Client struct {
	srv *server.Client
}

func New(srv *server.Client) *Client {
	return &Client{
		srv: srv,
	}
}

type DiscoverResult struct {
	AgentID  string
	Username string
	Hostname string
	System   string
	Address  string
	Created  time.Time
	LastSeen time.Time
}

func (c *Client) Discover(ctx context.Context) (map[string]DiscoverResult, error) {
	api := router.Handlers{
		proto.UpdateClientInfoSubject("<agent>"): func(ctx context.Context, params router.Params, data any) any {
			agentID, ok := params["agent"]
			if !ok {
				return nil
			}

			v, ok := data.(proto.UpdateClientInfo)
			if !ok {
				return nil
			}

			return DiscoverResult{
				AgentID:  agentID,
				Username: v.Username,
				Hostname: v.Hostname,
				System:   v.System,
				Address:  v.Address,
			}
		},
	}

	agentIDs, err := c.listAgentIDs(ctx)
	if err != nil {
		return nil, err
	}

	results := make(map[string]DiscoverResult)
	for _, agentID := range agentIDs {
		msgs, err := c.ListMessages(ctx, agentID, []string{
			proto.UpdateClientInfoSubject(agentID),
		})
		if err != nil {
			return nil, err
		}

		if len(msgs) == 0 {
			results[agentID] = DiscoverResult{
				AgentID: agentID,
				Created: time.Time{}, // TODO: grab timestamp from the stream!
			}
			continue
		}

		for _, msg := range msgs {
			handler, params, ok := api.Match(msg.Subject)
			if !ok {
				continue
			}

			data, err := proto.Unmarshal(msg.Data())
			if err != nil {
				continue
			}

			res := handler(ctx, params, data)
			if res == nil {
				continue
			}

			result := res.(DiscoverResult)
			results[agentID] = DiscoverResult{
				AgentID:  result.AgentID,
				Username: result.Username,
				Hostname: result.Hostname,
				System:   result.System,
				Address:  result.Address,
				Created:  time.Time{}, // TODO: grab timestamp from the stream!
				LastSeen: msg.Received,
			}
		}
	}

	return results, nil
}

func (c *Client) StartWorker(ctx context.Context, agentID, workerID, command string, args, env []string) error {
	b, err := proto.Marshal(proto.StartWorkerRequest{
		Command: command,
		Args:    args,
		Env:     env,
	})
	if err != nil {
		return err
	}

	err = c.publishMsg(
		ctx,
		StreamName(agentID),
		&nats.Msg{
			Subject: proto.StartWorkerSubject(agentID, workerID),
			Data:    b,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) StopWorker(ctx context.Context, agentID, workerID string) error {
	b, err := proto.Marshal(proto.StopWorkerRequest{})
	if err != nil {
		return err
	}

	err = c.publishMsg(
		ctx,
		StreamName(agentID),
		&nats.Msg{
			Subject: proto.StopWorkerSubject(agentID, workerID),
			Data:    b,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) WriteWorkerStdin(ctx context.Context, agentID, workerID string) error {
	b, err := proto.Marshal(proto.UpdateWorkerStdio{})
	if err != nil {
		return err
	}

	err = c.publishMsg(
		ctx,
		StreamName(agentID),
		&nats.Msg{
			Subject: proto.WriteWorkerStdinSubject(agentID, workerID),
			Data:    b,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ListJobs(ctx context.Context, agentID string) (map[string]Job, error) {
	jobs := make(map[string]Job)
	jobsMsgIDRef := make(map[string]string)

	api := router.Handlers{
		proto.StartWorkerSubject("<agent>", "<worker>"): func(ctx context.Context, params router.Params, data any) any {
			workerID, ok := params["worker"]
			if !ok {
				return nil
			}

			v, ok := data.(proto.StartWorkerRequest)
			if !ok {
				return nil
			}

			return Job{
				ID:      workerID,
				AgentID: agentID,
				Command: v.Command,
				Args:    strings.Join(v.Args, " "),
				Status:  JobStatusPending,
			}
		},
		proto.StopWorkerSubject("<agent>", "<worker>"): func(ctx context.Context, params router.Params, data any) any {
			workerID, ok := params["worker"]
			if !ok {
				return nil
			}

			job, ok := jobs[workerID]
			if !ok {
				return nil
			}

			_, ok = data.(proto.StopWorkerRequest)
			if !ok {
				return nil
			}

			job.Status = JobStatusCancelling

			return job
		},
		proto.UpdateWorkerStatusSubject("<agent>", "<worker>"): func(ctx context.Context, params router.Params, data any) any {
			workerID, ok := params["worker"]
			if !ok {
				return nil
			}

			job, ok := jobs[workerID]
			if !ok {
				return nil
			}

			v, ok := data.(proto.UpdateWorkerStatus)
			if !ok {
				return nil
			}

			switch v.Status {
			case 0:
				job.Status = JobStatusFinished
			case 130:
				job.Status = JobStatusCancelled
			default:
				job.Status = JobStatusFailed
			}

			return job
		},
		proto.ReplyMessageSubject("<agent>", "<message>"): func(ctx context.Context, params router.Params, data any) any {
			msgID, ok := params["message"]
			if !ok {
				return nil
			}

			jobRef, ok := jobsMsgIDRef[msgID]
			if !ok {
				return nil
			}

			job, ok := jobs[jobRef]
			if !ok {
				return nil
			}

			switch v := data.(type) {
			case proto.StartWorkerResponse:
				if v.Error != nil {
					job.Error = v.Error
					job.Status = JobStatusFailed
				} else {
					job.Status = JobStatusRunning
				}
				return job

			case proto.StopWorkerResponse:
				if v.Error != nil {
					job.Error = v.Error
					job.Status = JobStatusFailed
				} else {
					job.Status = JobStatusCancelled
				}
				return job
			}

			return nil
		},
	}

	msgs, err := c.ListMessages(ctx, agentID, []string{
		proto.StartWorkerSubject(agentID, "*"),
		proto.StopWorkerSubject(agentID, "*"),
		proto.UpdateWorkerStatusSubject(agentID, "*"),
		proto.ReplyMessageSubject(agentID, "*"),
	})
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		handler, params, ok := api.Match(msg.Subject)
		if !ok {
			continue
		}

		data, err := proto.Unmarshal(msg.Data())
		if err != nil {
			continue
		}

		res := handler(ctx, params, data)
		if res == nil {
			continue
		}

		job := res.(Job)
		jobs[job.ID] = job
		jobsMsgIDRef[msg.ID] = job.ID
	}

	return jobs, nil
}

func (c *Client) ListAllJobs(ctx context.Context) (map[string]Job, error) {
	agentIDs, err := c.listAgentIDs(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]Job)
	for _, agentID := range agentIDs {
		jobs, err := c.ListJobs(ctx, agentID)
		if err != nil {
			return nil, err
		}

		maps.Copy(result, jobs)
	}

	return result, nil
}

func (c *Client) GetJob(ctx context.Context, jobID string) (Job, error) {
	jobs, err := c.ListAllJobs(ctx)
	if err != nil {
		return Job{}, err
	}

	job, ok := jobs[jobID]
	if !ok {
		return Job{}, ErrJobNotFound
	}

	return job, nil
}

func (c *Client) CreateStorage(ctx context.Context, name, description string) error {
	_, err := c.srv.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{
		Bucket:      name,
		Description: description,
	})
	if err != nil {
		return &errorApi{err}
	}
	return nil
}

func (c *Client) DeleteStorage(ctx context.Context, name string) error {
	err := c.srv.DeleteObjectStore(ctx, name)
	if err != nil {
		return &errorApi{err}
	}
	return nil
}

func (c *Client) ListStorage(ctx context.Context) ([]*Storage, error) {
	var result []*Storage
	for name := range c.srv.ObjectStoreNames(ctx).Name() {
		store, err := c.srv.ObjectStore(ctx, name)
		if err != nil {
			return nil, &errorApi{err}
		}

		s, err := NewStorage(ctx, store)
		if err != nil {
			return nil, err
		}

		result = append(result, s)
	}
	return result, nil
}

func (c *Client) GetStorage(ctx context.Context, name string) (*Storage, error) {
	store, err := c.srv.ObjectStore(ctx, name)
	if err != nil {
		return nil, &errorApi{err}
	}

	s, err := NewStorage(ctx, store)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) ListMessages(ctx context.Context, agentID string, subjects []string) ([]Message, error) {
	consumer, err := c.createConsumer(ctx, agentID, subjects)
	if err != nil {
		return nil, err
	}

	var msgs []Message
	for {
		batch, err := consumer.FetchNoWait(350)
		if err != nil {
			return nil, err
		}

		var cnt int
		for msg := range batch.Messages() {
			if msg == nil {
				break
			}
			cnt++

			err := msg.Ack()
			if err != nil {
				return nil, err
			}

			meta, err := msg.Metadata()
			if err != nil {
				return nil, err
			}

			msgID := msg.Headers().Get(nats.MsgIdHdr)

			msgs = append(msgs, Message{
				ID:       msgID,
				Subject:  msg.Subject(),
				AgentID:  agentID,
				Sent:     time.Time{}, // TODO: extract from the message headers!
				Received: meta.Timestamp,
				msg:      msg,
			})
		}

		err = batch.Error()
		if err != nil {
			return nil, err
		}

		if cnt == 0 {
			break
		}
	}

	msgs = slices.SortedFunc(slices.Values(msgs), func(v1, v2 Message) int {
		if !v1.Sent.IsZero() && !v2.Sent.IsZero() {
			if v1.Sent.Before(v2.Sent) {
				return -1
			}
			if v1.Sent.After(v2.Sent) {
				return 1
			}
			return 0
		}
		if v1.Received.Before(v2.Received) {
			return -1
		}
		if v1.Received.After(v2.Received) {
			return 1
		}
		return 0
	})

	return msgs, nil
}

func (c *Client) publishMsg(ctx context.Context, stream string, msg *nats.Msg) error {
	if msg.Header == nil {
		msg.Header = make(nats.Header)
	}
	msg.Header.Set(nats.MsgIdHdr, nuid.Next())
	_, err := c.srv.PublishMsg(
		ctx,
		msg,
		jetstream.WithExpectStream(stream),
	)
	if err != nil {
		return &errorApi{err}
	}
	return nil
}

func (c *Client) listStreams(ctx context.Context) ([]string, error) {
	var names []string
	for stream := range c.srv.ListStreams(ctx).Info() {
		if stream == nil {
			break
		}
		names = append(names, stream.Config.Name)
	}
	return names, nil
}

func (c *Client) createConsumer(ctx context.Context, agentID string, subjects []string) (jetstream.Consumer, error) {
	consumer, err := c.srv.CreateConsumer(ctx, StreamName(agentID), jetstream.ConsumerConfig{
		DeliverPolicy:  jetstream.DeliverAllPolicy,
		AckPolicy:      jetstream.AckExplicitPolicy,
		MaxAckPending:  1,
		FilterSubjects: subjects,
	})
	if err != nil {
		return nil, &errorApi{err}
	}
	return consumer, nil
}

func (c *Client) listAgentIDs(ctx context.Context) ([]string, error) {
	streams, err := c.listStreams(ctx)
	if err != nil {
		return nil, err
	}

	var agentIDs []string
	for _, stream := range streams {
		if !hasStreamPrefix(stream) {
			continue
		}
		agentID := trimStreamPrefix(stream)
		agentIDs = append(agentIDs, agentID)
	}

	return agentIDs, nil
}

const streamPrefix = "FJ_"

func StreamName(name string) string {
	return streamPrefix + name
}

func trimStreamPrefix(name string) string {
	return strings.TrimPrefix(name, streamPrefix)
}

func hasStreamPrefix(name string) bool {
	return strings.HasPrefix(name, streamPrefix)
}

const InboxPrefix = "_INBOX_"

func InboxName(name string) string {
	return InboxPrefix + name
}

func NewAgentPermissions(agentID string) jwt.Permissions {
	return jwt.Permissions{
		Pub: jwt.Permission{
			Allow: []string{
				proto.WriteWorkerStdoutSubject(agentID, "*"),
				proto.UpdateWorkerStatusSubject(agentID, "*"),
				proto.ReplyMessageSubject(agentID, "*"),
				proto.UpdateClientInfoSubject(agentID),

				fmt.Sprintf("$JS.ACK.FJ_%s.>", agentID),
				fmt.Sprintf("$JS.API.STREAM.INFO.FJ_%s", agentID),
				fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", agentID),
				fmt.Sprintf("$JS.API.STREAM.PURGE.OBJ_%s", agentID),
				fmt.Sprintf("$JS.API.CONSUMER.INFO.FJ_%s.*", agentID),
				fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.FJ_%s.*", agentID),
				fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.M.*", agentID, agentID),
				fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.>", agentID),
				fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", agentID),
				fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.>", agentID),
				fmt.Sprintf("$O.%s.M.*", agentID),
				fmt.Sprintf("$O.%s.C.*", agentID),
			},
		},
		Sub: jwt.Permission{
			Allow: []string{
				fmt.Sprintf("_INBOX_%s.>", agentID),
			},
		},
	}
}
