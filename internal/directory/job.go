package directory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type JobDirectory struct {
	Directory
}

func OpenJobDirectory(ctx context.Context, js jetstream.JetStream) (*JobDirectory, error) {
	dir, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
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

func (d *JobDirectory) Create(ctx context.Context, entry JobDirectoryEntry) (JobDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	rev, err := d.Directory.Create(ctx, formatKey(entry.AgentID, entry.ID), b, entry.ID)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *JobDirectory) Update(ctx context.Context, entry JobDirectoryEntry) (JobDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	rev, err := d.Directory.Update(ctx, formatKey(entry.AgentID, entry.ID), b, entry.Revision)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *JobDirectory) Get(ctx context.Context, key string) (JobDirectoryEntry, error) {
	v, err := d.Directory.Get(ctx, key)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	var entry JobDirectoryEntry
	err = json.Unmarshal(v.Value, &entry)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	entry.Revision = v.Revision

	return entry, nil
}

func (d *JobDirectory) List(ctx context.Context) ([]JobDirectoryEntry, error) {
	return d.list(ctx, formatKey("*", "*"))
}

func (d *JobDirectory) ListByAgentID(ctx context.Context, agentID string) ([]JobDirectoryEntry, error) {
	return d.list(ctx, formatKey(agentID, "*"))
}

func (d *JobDirectory) Delete(ctx context.Context, job JobDirectoryEntry) error {
	return d.Directory.Delete(ctx, formatKey(job.AgentID, job.ID), job.ID)
}

func (d *JobDirectory) list(ctx context.Context, key string) ([]JobDirectoryEntry, error) {
	blobs, err := d.Directory.List(ctx, key)
	if err != nil {
		return nil, err
	}

	entries := make([]JobDirectoryEntry, 0, len(blobs))
	for _, b := range blobs {
		var entry JobDirectoryEntry
		err := json.Unmarshal(b.Value, &entry)
		if err != nil {
			return nil, err
		}

		entry.Revision = b.Revision
		entries = append(entries, entry)
	}

	return entries, nil
}

type JobDirectoryEntry struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	WorkerID  string    `json:"worker_id"`
	GatewayID string    `json:"gateway_id"`
	Config    JobConfig `json:"config"`
	State     JobState  `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	Revision  uint64    `json:"-"`
}

type JobConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env"`
}

type JobState struct {
	Status    string    `json:"status"`
	Error     string    `json:"error"`
	UpdatedAt time.Time `json:"updated_at"`
}
