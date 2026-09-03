package daemon

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const idKeyPrefix = "id."

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

type AgentDirectory struct {
	Directory
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
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	GatewayID string           `json:"gateway_id"`
	Config    AgentBuildConfig `json:"config"`
	Revision  uint64           `json:"-"`
}

type AgentBuildConfig struct {
	OS      string            `json:"os"`
	Arch    string            `json:"arch"`
	UserJWT string            `json:"user_jwt"`
	UserKey string            `json:"user_key"`
	Extra   map[string]string `json:"extra"`
}

type AgentHostDirectory struct {
	Directory
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

type GatewayDirectory struct {
	Directory
}

func (d *GatewayDirectory) Create(ctx context.Context, entry GatewayDirectoryEntry) (GatewayDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

	rev, err := d.Directory.Create(ctx, entry.ID, b, entry.Name)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *GatewayDirectory) Update(ctx context.Context, entry GatewayDirectoryEntry) (GatewayDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

	rev, err := d.Directory.Update(ctx, entry.ID, b, entry.Revision)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *GatewayDirectory) Get(ctx context.Context, key string) (GatewayDirectoryEntry, error) {
	v, err := d.Directory.Get(ctx, key)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

	var entry GatewayDirectoryEntry
	err = json.Unmarshal(v.Value, &entry)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

	entry.Revision = v.Revision

	return entry, nil
}

func (d *GatewayDirectory) List(ctx context.Context) ([]GatewayDirectoryEntry, error) {
	blobs, err := d.Directory.List(ctx, formatKey("*"))
	if err != nil {
		return nil, err
	}

	entries := make([]GatewayDirectoryEntry, 0, len(blobs))
	for _, b := range blobs {
		var entry GatewayDirectoryEntry
		err := json.Unmarshal(b.Value, &entry)
		if err != nil {
			return nil, err
		}

		entry.Revision = b.Revision
		entries = append(entries, entry)
	}

	return entries, nil
}

func (d *GatewayDirectory) Delete(ctx context.Context, gateway GatewayDirectoryEntry) error {
	return d.Directory.Delete(ctx, gateway.ID, gateway.Name)
}

type GatewayDirectoryEntry struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Config      GatewayConfig `json:"config"`
	Revision    uint64        `json:"-"`
}

type GatewayConfig struct {
	UserJWT string            `json:"user_jwt"`
	UserKey string            `json:"user_key"`
	Extra   map[string]string `json:"extra"`
}

type JobDirectory struct {
	Directory
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
