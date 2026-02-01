package agent

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/clients/server"
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

func (c *Client) Register(ctx context.Context, agentID string) (err error) {
	streamName := StreamName(agentID)
	_, err = c.srv.CreateStream(ctx, jetstream.StreamConfig{
		Name: streamName,
		Subjects: []string{
			proto.StartWorkerSubject(agentID, "*"),
			proto.StopWorkerSubject(agentID, "*"),
			proto.WriteWorkerStdinSubject(agentID, "*"),
			proto.WriteWorkerStdoutSubject(agentID, "*"),
			proto.UpdateWorkerStatusSubject(agentID, "*"),
			proto.ReplyMessageSubject(agentID, "*"),
			proto.UpdateClientInfoSubject(agentID),
		},
	})
	if err != nil {
		return &errorApi{err}
	}
	defer func() {
		if err == nil {
			return
		}
		_ = c.srv.DeleteStream(ctx, streamName)
	}()

	consumerName := agentID
	_, err = c.srv.CreateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: jetstream.DeliverLastPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
		FilterSubjects: []string{
			proto.StartWorkerSubject(agentID, "*"),
			proto.StopWorkerSubject(agentID, "*"),
			proto.WriteWorkerStdinSubject(agentID, "*"),
		},
	})
	if err != nil {
		return &errorApi{err}
	}
	defer func() {
		if err == nil {
			return
		}
		_ = c.srv.DeleteConsumer(ctx, streamName, consumerName)
	}()

	storeName := agentID
	storeDesc := fmt.Sprintf("Agent %s storage", agentID)
	_, err = c.srv.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{
		Bucket:      storeName,
		Description: storeDesc,
	})
	if err != nil {
		return &errorApi{err}
	}
	defer func() {
		if err == nil {
			return
		}
		_ = c.srv.DeleteObjectStore(ctx, storeName)
	}()

	return nil
}

func (c *Client) Unregister(ctx context.Context, agentID string) error {
	storeName := agentID
	err := c.srv.DeleteObjectStore(ctx, storeName)
	if err != nil && !errors.Is(err, nats.ErrBucketNotFound) {
		return &errorApi{err}
	}

	streamName := StreamName(agentID)
	consumerName := agentID
	err = c.srv.DeleteConsumer(ctx, streamName, consumerName)
	if err != nil && !errors.Is(err, nats.ErrConsumerNotFound) {
		return &errorApi{err}
	}

	err = c.srv.DeleteStream(ctx, streamName)
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return &errorApi{err}
	}

	return nil
}

type DiscoverResult struct {
	AgentID  string
	Username string
	Hostname string
	System   string
	Address  string
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
		}, 1, -1)
		if err != nil {
			return nil, err
		}

		if len(msgs) == 0 {
			results[agentID] = DiscoverResult{
				AgentID: agentID,
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

type contextKey string

var messageKey contextKey = "foojank:message"

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

			msg := ctx.Value(messageKey).(Message)

			return Job{
				ID:      workerID,
				AgentID: agentID,
				Command: v.Command,
				Args:    strings.Join(v.Args, " "),
				Status:  JobStatusPending,
				Created: msg.Received,
				Updated: msg.Received,
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

			msg := ctx.Value(messageKey).(Message)

			job.Status = JobStatusCancelling
			job.Updated = msg.Received

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

			msg := ctx.Value(messageKey).(Message)

			switch v.Status {
			case 0:
				job.Status = JobStatusFinished
			case 130:
				job.Status = JobStatusCancelled
			default:
				job.Status = JobStatusFailed
			}

			job.Updated = msg.Received

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

			msg := ctx.Value(messageKey).(Message)

			switch v := data.(type) {
			case proto.StartWorkerResponse:
				if v.Error != nil {
					job.Error = v.Error
					job.Status = JobStatusFailed
				} else {
					job.Status = JobStatusRunning
				}

			case proto.StopWorkerResponse:
				if v.Error != nil {
					job.Error = v.Error
					job.Status = JobStatusFailed
				} else {
					job.Status = JobStatusCancelled
				}

			default:
				return nil
			}

			job.Updated = msg.Received

			return job
		},
	}

	msgs, err := c.ListMessages(ctx, agentID, []string{
		proto.StartWorkerSubject(agentID, "*"),
		proto.StopWorkerSubject(agentID, "*"),
		proto.UpdateWorkerStatusSubject(agentID, "*"),
		proto.ReplyMessageSubject(agentID, "*"),
	}, 1, -1)
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

		res := handler(context.WithValue(ctx, messageKey, msg), params, data)
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

func (c *Client) ListMessages(
	ctx context.Context,
	agentID string,
	subjects []string,
	startSeq uint64,
	limit int,
) ([]Message, error) {
	consumer, err := c.srv.CreateConsumer(ctx, StreamName(agentID), jetstream.ConsumerConfig{
		DeliverPolicy:  jetstream.DeliverByStartSequencePolicy,
		AckPolicy:      jetstream.AckNonePolicy,
		MaxAckPending:  1,
		FilterSubjects: subjects,
		OptStartSeq:    startSeq,
	})
	if err != nil {
		return nil, &errorApi{err}
	}

	if limit <= 0 {
		limit = 50
	}

	var msgs []Message
	for {
		batch, err := consumer.FetchNoWait(limit - len(msgs))
		if err != nil {
			return nil, err
		}

		cnt := 0
		for msg := range batch.Messages() {
			if msg == nil {
				break
			}

			meta, err := msg.Metadata()
			if err != nil {
				return nil, err
			}

			msgID := msg.Headers().Get(nats.MsgIdHdr)
			msgs = append(msgs, Message{
				ID:       msgID,
				Seq:      meta.Sequence.Stream,
				Subject:  msg.Subject(),
				AgentID:  agentID,
				Sent:     time.Time{}, // TODO: extract from the message headers!
				Received: meta.Timestamp,
				msg:      msg,
			})
			cnt++
		}

		err = batch.Error()
		if err != nil {
			return nil, err
		}

		if cnt == 0 || len(msgs) == limit {
			break
		}
	}

	return msgs, nil
}

func (c *Client) StreamMessages(
	ctx context.Context,
	agentID string,
	subjects []string,
	startSeq uint64,
	outputCh chan<- Message,
) error {
	consumer, err := c.srv.CreateConsumer(ctx, StreamName(agentID), jetstream.ConsumerConfig{
		DeliverPolicy:  jetstream.DeliverByStartSequencePolicy,
		AckPolicy:      jetstream.AckNonePolicy,
		MaxAckPending:  1,
		FilterSubjects: subjects,
		OptStartSeq:    startSeq,
	})
	if err != nil {
		return &errorApi{err}
	}

	msgs, err := consumer.Messages()
	if err != nil {
		return err
	}

	for {
		msg, err := msgs.Next(jetstream.NextContext(ctx))
		if err != nil {
			if errors.Is(err, jetstream.ErrMsgIteratorClosed) {
				return nil
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}

		meta, err := msg.Metadata()
		if err != nil {
			return err
		}

		msgID := msg.Headers().Get(nats.MsgIdHdr)

		select {
		case outputCh <- Message{
			ID:       msgID,
			Seq:      meta.Sequence.Stream,
			Subject:  msg.Subject(),
			AgentID:  agentID,
			Sent:     time.Time{}, // TODO: extract from the message headers!
			Received: meta.Timestamp,
			msg:      msg,
		}:
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Client) StreamWorkerStdio(
	ctx context.Context,
	agentID string,
	workerID string,
	startSeq uint64,
	outputCh chan<- []byte,
) error {
	var msgCh = make(chan Message)
	var errCh = make(chan error, 1)
	go func() {
		err := c.StreamMessages(
			ctx,
			agentID,
			[]string{
				proto.WriteWorkerStdinSubject(agentID, workerID),
				proto.WriteWorkerStdoutSubject(agentID, workerID),
			},
			startSeq,
			msgCh,
		)
		errCh <- err
	}()

	for {
		// Select does not need to monitor ctx.
		// The loop is terminated when errCh emits a value.
		// This ensures that the go routine has stopped and so cannot leak.
		select {
		case msg := <-msgCh:
			data, err := proto.Unmarshal(msg.Data())
			if err != nil {
				continue
			}

			v := data.(proto.UpdateWorkerStdio)
			outputCh <- v.Data

		case err := <-errCh:
			return err
		}
	}
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
