package vessel

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/router"
	"github.com/foohq/foojank/proto"
)

const (
	SubjectApiPrefix             = "FJ.API"
	SubjectApiWorkerStartT       = SubjectApiPrefix + "." + "WORKER.START.%s.%s"
	SubjectApiWorkerStopT        = SubjectApiPrefix + "." + "WORKER.STOP.%s.%s"
	SubjectApiWorkerWriteStdinT  = SubjectApiPrefix + "." + "WORKER.WRITE.STDIN.%s.%s"
	SubjectApiWorkerWriteStdoutT = SubjectApiPrefix + "." + "WORKER.WRITE.STDOUT.%s.%s"
	SubjectApiWorkerStatusT      = SubjectApiPrefix + "." + "WORKER.STATUS.%s.%s"
	SubjectApiConnInfoT          = SubjectApiPrefix + "." + "CONNECTION.INFO.%s"
	SubjectApiReplyT             = SubjectApiPrefix + "." + "REPLY.%s.%s"
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
	ID       string
	Username string
	Hostname string
	System   string
	Address  string
	Created  time.Time
	LastSeen time.Time
}

func (c *Client) Discover(ctx context.Context) ([]DiscoverResult, error) {
	api := router.Handlers{
		"FJ.API.CONNECTION.INFO.<agent>": func(ctx context.Context, params router.Params, data any) any {
			agentID, ok := params["agent"]
			if !ok {
				return nil
			}

			v, ok := data.(proto.UpdateClientInfo)
			if !ok {
				return nil
			}

			return DiscoverResult{
				ID:       agentID,
				Username: v.Username,
				Hostname: v.Hostname,
				System:   v.System,
				Address:  v.Address,
			}
		},
	}

	agentIDs, err := c.ListAgentIDs(ctx)
	if err != nil {
		return nil, err
	}

	var results []DiscoverResult
	for _, agentID := range agentIDs {
		msgs, err := c.ListMessages(ctx, agentID)
		if err != nil {
			return nil, err
		}

		if len(msgs) == 0 {
			results = append(results, DiscoverResult{
				ID:      agentID,
				Created: time.Time{}, // TODO: grab timestamp from the stream!
			})
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
			results = append(results, DiscoverResult{
				ID:       result.ID,
				Username: result.Username,
				Hostname: result.Hostname,
				System:   result.System,
				Address:  result.Address,
				Created:  time.Time{}, // TODO: grab timestamp from the stream!
				LastSeen: msg.Received,
			})
		}
	}

	return results, nil
}

func (c *Client) StartWorker(ctx context.Context, agentID, workerID, file string, args, env []string) error {
	b, err := proto.Marshal(proto.StartWorkerRequest{
		File: file,
		Args: args,
		Env:  env,
	})
	if err != nil {
		return err
	}

	err = c.srv.Publish(ctx, &nats.Msg{
		Subject: fmt.Sprintf(SubjectApiWorkerStartT, agentID, workerID),
		Data:    b,
	})
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

	err = c.srv.Publish(ctx, &nats.Msg{
		Subject: fmt.Sprintf(SubjectApiWorkerStopT, agentID, workerID),
		Data:    b,
	})
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

	err = c.srv.Publish(ctx, &nats.Msg{
		Subject: fmt.Sprintf(SubjectApiWorkerWriteStdinT, agentID, workerID),
		Data:    b,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ListAgentIDs(ctx context.Context) ([]string, error) {
	streams, err := c.srv.ListStreams(ctx)
	if err != nil {
		return nil, err
	}

	var agentIDs []string
	for _, stream := range streams {
		if !strings.HasPrefix(stream, StreamPrefix) {
			continue
		}
		agentID := strings.TrimPrefix(stream, StreamPrefix)
		agentIDs = append(agentIDs, agentID)
	}

	return agentIDs, nil
}

func (c *Client) CreateConsumer(ctx context.Context, agentID string) (jetstream.Consumer, error) {
	return c.srv.CreateConsumer(ctx, StreamName(agentID))
}

func (c *Client) ListJobs(ctx context.Context, agentID string) (map[string]*Job, error) {
	jobs := make(map[string]*Job)
	jobsMsgID := make(map[string]*Job)

	api := router.Handlers{
		"FJ.API.WORKER.START.<agent>.<worker>": func(ctx context.Context, params router.Params, data any) any {
			workerID, ok := params["worker"]
			if !ok {
				return nil
			}

			v, ok := data.(proto.StartWorkerRequest)
			if !ok {
				return nil
			}

			return &Job{
				id:      workerID,
				agentID: agentID,
				info: JobInfo{
					File:   v.File,
					Args:   strings.Join(v.Args, " "),
					Status: JobStatusPending,
				},
			}
		},
		"FJ.API.WORKER.STOP.<agent>.<worker>": func(ctx context.Context, params router.Params, data any) any {
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

			job.info.Status = JobStatusCancelling

			return job
		},
		"FJ.API.WORKER.STATUS.<agent>.<worker>": func(ctx context.Context, params router.Params, data any) any {
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
				job.info.Status = JobStatusFinished
			case 130:
				job.info.Status = JobStatusCancelled
			default:
				job.info.Status = JobStatusFailed
			}

			return job
		},
		"FJ.API.REPLY.<agent>.<message>": func(ctx context.Context, params router.Params, data any) any {
			msgID, ok := params["message"]
			if !ok {
				return nil
			}

			job, ok := jobsMsgID[msgID]
			if !ok {
				return nil
			}

			switch v := data.(type) {
			case proto.StartWorkerResponse:
				if v.Error != nil {
					job.info.Error = v.Error
					job.info.Status = JobStatusFailed
				} else {
					job.info.Status = JobStatusRunning
				}
				return job

			case proto.StopWorkerResponse:
				if v.Error != nil {
					job.info.Error = v.Error
					job.info.Status = JobStatusFailed
				} else {
					job.info.Status = JobStatusCancelled
				}
				return job
			}

			return nil
		},
	}

	msgs, err := c.ListMessages(ctx, agentID)
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

		job := res.(*Job)
		jobs[job.id] = job
	}

	return jobs, nil
}

func (c *Client) ListAllJobs(ctx context.Context) (map[string]*Job, error) {
	agentIDs, err := c.ListAgentIDs(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*Job)
	for _, agentID := range agentIDs {
		jobs, err := c.ListJobs(ctx, agentID)
		if err != nil {
			return nil, err
		}

		maps.Copy(result, jobs)
	}

	return result, nil
}

type Message struct {
	msg      jetstream.Msg
	ID       string
	Subject  string
	AgentID  string
	Sent     time.Time
	Received time.Time
}

func (m *Message) Data() []byte {
	return m.msg.Data()
}

func (c *Client) ListMessages(ctx context.Context, agentID string) ([]*Message, error) {
	consumer, err := c.CreateConsumer(ctx, agentID)
	if err != nil {
		return nil, err
	}

	var msgs []*Message

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

			msgs = append(msgs, &Message{
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

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].ID >= msgs[j].ID
	})

	return msgs, nil
}

type Job struct {
	id      string
	agentID string
	info    JobInfo
}

func (j *Job) ID() string {
	return j.id
}

func (j *Job) AgentID() string {
	return j.agentID
}

func (j *Job) Info() JobInfo {
	return j.info
}

const (
	JobStatusPending    = "Pending"
	JobStatusRunning    = "Running"
	JobStatusCancelling = "Cancelling"
	JobStatusCancelled  = "Cancelled"
	JobStatusFinished   = "Finished"
	JobStatusFailed     = "Failed"
)

type JobInfo struct {
	File   string
	Args   string
	Status string
	Error  error
}

const StreamPrefix = "FJ_"

func StreamName(name string) string {
	return StreamPrefix + name
}

const InboxPrefix = "_INBOX_"

func InboxName(name string) string {
	return InboxPrefix + name
}
