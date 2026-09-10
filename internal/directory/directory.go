package directory

import (
	"context"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

const idKeyPrefix = "id."

const (
	agentDirectoryName     = "agents"
	agentHostDirectoryName = "agent-hosts"
	gatewayDirectoryName   = "gateways"
	jobsDirectoryName      = "jobs"
	userDirectoryName      = "users"
)

func formatKey(parts ...string) string {
	return strings.Join(parts, ".")
}

type Directory struct {
	store jetstream.KeyValue
}

func (d *Directory) Create(ctx context.Context, id string, value []byte, keys ...string) (revision uint64, err error) {
	id = idKeyPrefix + strings.ToLower(id)
	revision, err = d.store.Create(ctx, id, value)
	if err != nil {
		return 0, err
	}

	created := make([]string, 0, len(keys))
	defer func() {
		if err == nil {
			return
		}
		_ = d.store.Delete(ctx, id)
		for _, k := range created {
			_ = d.store.Delete(ctx, k)
		}
	}()

	for _, key := range keys {
		key = strings.ToLower(key)
		_, err = d.store.Create(ctx, key, []byte(id))
		if err != nil {
			return 0, err
		}
		created = append(created, key)
	}

	return revision, nil
}

func (d *Directory) Update(ctx context.Context, id string, value []byte, revision uint64) (uint64, error) {
	id = idKeyPrefix + strings.ToLower(id)
	revision, err := d.store.Update(ctx, id, value, revision)
	if err != nil {
		return 0, err
	}

	return revision, nil
}

func (d *Directory) Delete(ctx context.Context, id string, keys ...string) error {
	id = idKeyPrefix + strings.ToLower(id)
	err := d.store.Purge(ctx, id)
	if err != nil {
		return err
	}

	for _, key := range keys {
		key = strings.ToLower(key)
		err := d.store.Purge(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Directory) Get(ctx context.Context, key string) (DirectoryEntry, error) {
	key = strings.ToLower(key)

	// Try as a direct id key first.
	v, err := d.store.Get(ctx, idKeyPrefix+key)
	if err == nil {
		return DirectoryEntry{
			Value:    v.Value(),
			Revision: v.Revision(),
		}, nil
	}

	// Try as a reference key.
	v, err = d.store.Get(ctx, key)
	if err != nil {
		return DirectoryEntry{}, err
	}

	// Resolve reference to id value.
	v2, err := d.store.Get(ctx, string(v.Value()))
	if err != nil {
		return DirectoryEntry{}, err
	}

	return DirectoryEntry{
		Value:    v2.Value(),
		Revision: v2.Revision(),
	}, nil
}

func (d *Directory) List(ctx context.Context, key string) ([]DirectoryEntry, error) {
	list, err := d.store.ListKeysFiltered(ctx, idKeyPrefix+strings.ToLower(key))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = list.Stop()
	}()

	seen := make(map[string]struct{})
	for key := range list.Keys() {
		seen[key] = struct{}{}
	}

	result := make([]DirectoryEntry, 0, len(seen))
	for key := range seen {
		v, err := d.store.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		result = append(result, DirectoryEntry{
			Value:    v.Value(),
			Revision: v.Revision(),
		})
	}

	return result, nil
}

type DirectoryEntry struct {
	Value    []byte
	Revision uint64
}
