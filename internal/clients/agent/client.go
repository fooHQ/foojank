package agent

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"

	"github.com/foohq/foojank/internal/clients/server"
)

const (
	agentDirectoryName   = "agent-directory"
	gatewayDirectoryName = "gateway-directory"
)

const (
	StreamType        = "Foojank-Stream-Type"
	StreamTypeAgent   = "agent"
	StreamTypeGateway = "gateway"
)

type Client struct {
	srv *server.Client
}

func New(srv *server.Client) *Client {
	return &Client{
		srv: srv,
	}
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
		s, err := c.GetStorage(ctx, name)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (c *Client) GetStorage(ctx context.Context, name string) (*Storage, error) {
	agentDir, err := c.openDirectory(ctx, agentDirectoryName)
	if err != nil {
		// If the bucket does not exist, return an empty result.
		// Bucket does not exist only before the first agent is registered.
		if errors.Is(err, jetstream.ErrBucketNotFound) {
			return nil, nil
		}
		return nil, &errorApi{err}
	}

	store, err := c.srv.ObjectStore(ctx, name)
	if err != nil {
		return nil, &errorApi{err}
	}

	var storageName string
	v, err := agentDir.Get(ctx, name)
	if err == nil {
		storageName = v
	} else if !errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, &errorApi{err}
	}

	s, err := NewStorage(ctx, storageName, store)
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
	consumer, err := c.srv.CreateConsumer(ctx, agentID, jetstream.ConsumerConfig{
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

func (c *Client) IsAgentID(agentID string) bool {
	return nkeys.IsValidPublicUserKey(agentID)
}

func (c *Client) GetAgentID(ctx context.Context, name string) (string, error) {
	if c.IsAgentID(name) {
		return name, nil
	}

	agentDir, err := c.openDirectory(ctx, agentDirectoryName)
	if err != nil {
		// If the bucket does not exist, return an empty result.
		// Bucket does not exist only before the first agent is registered.
		if errors.Is(err, jetstream.ErrBucketNotFound) {
			return "", ErrAgentNotFound
		}
		return "", &errorApi{err}
	}

	v, err := agentDir.Get(ctx, name)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return "", ErrAgentNotFound
		}
		return "", &errorApi{err}
	}

	return v, nil
}

func (c *Client) GetStorageName(ctx context.Context, name string) (string, error) {
	return c.GetAgentID(ctx, name)
}

func (c *Client) openDirectory(ctx context.Context, name string) (*AgentDirectory, error) {
	dir, err := c.srv.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: name,
	})
	if err != nil && !errors.Is(err, jetstream.ErrBucketExists) {
		return nil, err
	}
	return &AgentDirectory{
		store: dir,
	}, nil
}
