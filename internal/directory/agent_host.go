package directory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type AgentHostDirectory struct {
	Directory
}

func OpenAgentHostDirectory(ctx context.Context, js jetstream.JetStream) (*AgentHostDirectory, error) {
	dir, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
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

func (d *AgentHostDirectory) Create(ctx context.Context, entry AgentHostDirectoryEntry) (AgentHostDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

	rev, err := d.Directory.Create(ctx, entry.AgentID, b)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *AgentHostDirectory) Update(ctx context.Context, entry AgentHostDirectoryEntry) (AgentHostDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

	rev, err := d.Directory.Update(ctx, entry.AgentID, b, entry.Revision)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *AgentHostDirectory) Get(ctx context.Context, key string) (AgentHostDirectoryEntry, error) {
	v, err := d.Directory.Get(ctx, key)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

	var entry AgentHostDirectoryEntry
	err = json.Unmarshal(v.Value, &entry)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

	entry.Revision = v.Revision

	return entry, nil
}

func (d *AgentHostDirectory) List(ctx context.Context) ([]AgentHostDirectoryEntry, error) {
	blobs, err := d.Directory.List(ctx, formatKey("*"))
	if err != nil {
		return nil, err
	}

	entries := make([]AgentHostDirectoryEntry, 0, len(blobs))
	for _, b := range blobs {
		var entry AgentHostDirectoryEntry
		err := json.Unmarshal(b.Value, &entry)
		if err != nil {
			return nil, err
		}

		entry.Revision = b.Revision
		entries = append(entries, entry)
	}

	return entries, nil
}

func (d *AgentHostDirectory) Delete(ctx context.Context, agent AgentHostDirectoryEntry) error {
	return d.Directory.Delete(ctx, agent.AgentID)
}

type AgentHostDirectoryEntry struct {
	AgentID    string    `json:"agent_id"`
	Username   string    `json:"username"`
	Hostname   string    `json:"hostname"`
	System     string    `json:"system"`
	Address    string    `json:"address"`
	LastUpdate time.Time `json:"last_update"`
	Revision   uint64    `json:"-"`
}
