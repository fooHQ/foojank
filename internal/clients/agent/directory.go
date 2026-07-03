package agent

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

type AgentDirectory struct {
	store jetstream.KeyValue
}

func (d *AgentDirectory) Create(ctx context.Context, agentID string, name string) (err error) {
	_, err = d.store.Create(ctx, agentID, []byte(name))
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			return
		}
		_ = d.store.Delete(ctx, agentID)
	}()

	_, err = d.store.Create(ctx, name, []byte(agentID))
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			return
		}
		_ = d.store.Delete(ctx, name)
	}()

	return nil
}

func (d *AgentDirectory) Delete(ctx context.Context, key string) error {
	v, err := d.store.Get(ctx, key)
	if err != nil {
		return err
	}

	err1 := d.store.Delete(ctx, key)
	err2 := d.store.Delete(ctx, string(v.Value()))

	switch {
	case err1 != nil:
		return err1
	case err2 != nil:
		return err2
	default:
		return nil
	}
}

func (d *AgentDirectory) Get(ctx context.Context, key string) (string, error) {
	v, err := d.store.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(v.Value()), nil
}
