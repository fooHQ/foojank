package directory

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
)

type AgentDirectory struct {
	Directory
}

func OpenAgentDirectory(ctx context.Context, js jetstream.JetStream) (*AgentDirectory, error) {
	dir, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
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

func (d *AgentDirectory) Create(ctx context.Context, entry AgentDirectoryEntry) (AgentDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

	rev, err := d.Directory.Create(ctx, entry.ID, b, entry.Name)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *AgentDirectory) Update(ctx context.Context, entry AgentDirectoryEntry) (AgentDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

	rev, err := d.Directory.Update(ctx, entry.ID, b, entry.Revision)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *AgentDirectory) Get(ctx context.Context, key string) (AgentDirectoryEntry, error) {
	v, err := d.Directory.Get(ctx, key)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

	var entry AgentDirectoryEntry
	err = json.Unmarshal(v.Value, &entry)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

	entry.Revision = v.Revision

	return entry, nil
}

func (d *AgentDirectory) List(ctx context.Context) ([]AgentDirectoryEntry, error) {
	blobs, err := d.Directory.List(ctx, formatKey("*"))
	if err != nil {
		return nil, err
	}

	entries := make([]AgentDirectoryEntry, 0, len(blobs))
	for _, b := range blobs {
		var entry AgentDirectoryEntry
		err := json.Unmarshal(b.Value, &entry)
		if err != nil {
			return nil, err
		}

		entry.Revision = b.Revision
		entries = append(entries, entry)
	}

	return entries, nil
}

func (d *AgentDirectory) Delete(ctx context.Context, agent AgentDirectoryEntry) error {
	return d.Directory.Delete(ctx, agent.ID, agent.Name)
}

type AgentDirectoryEntry struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	GatewayID   string           `json:"gateway_id"`
	Config      AgentBuildConfig `json:"config"`
	Revision    uint64           `json:"-"`
}

type AgentBuildConfig struct {
	OS      string            `json:"os"`
	Arch    string            `json:"arch"`
	UserJWT string            `json:"user_jwt"`
	UserKey string            `json:"user_key"`
	Extra   map[string]string `json:"extra"`
}
