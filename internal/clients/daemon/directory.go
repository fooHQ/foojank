package daemon

import (
	"context"
	"encoding/json"
	"errors"
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

func (d *Directory) Create(ctx context.Context, id string, value []byte, keys ...string) (err error) {
	id = idKeyPrefix + strings.ToLower(id)
	_, err = d.store.Create(ctx, id, value)
	if err != nil {
		return err
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
			return err
		}
		created = append(created, key)
	}

	return nil
}

func (d *Directory) Delete(ctx context.Context, key string) error {
	key = strings.ToLower(key)

	// Try deleting as an id key.
	id := idKeyPrefix + key
	err := d.store.Delete(ctx, id)
	if err == nil {
		return nil
	}

	// Try as a reference key, also clean up the id entry.
	v, err := d.store.Get(ctx, key)
	if err != nil {
		return err
	}

	err1 := d.store.Delete(ctx, key)
	_ = d.store.Delete(ctx, string(v.Value()))

	return err1
}

func (d *Directory) Get(ctx context.Context, key string) ([]byte, error) {
	key = strings.ToLower(key)

	// Try as a direct id key first.
	v, err := d.store.Get(ctx, idKeyPrefix+key)
	if err == nil {
		return v.Value(), nil
	}

	// Try as a reference key.
	v, err = d.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Resolve reference to id value.
	v2, err := d.store.Get(ctx, string(v.Value()))
	if err != nil {
		return nil, err
	}

	return v2.Value(), nil
}

func (d *Directory) List(ctx context.Context, key string) ([][]byte, error) {
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

	result := make([][]byte, 0, len(seen))
	for key := range seen {
		v, err := d.store.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		result = append(result, v.Value())
	}

	return result, nil
}

type AgentDirectory struct {
	Directory
}

func (d *AgentDirectory) Create(ctx context.Context, entry AgentDirectoryEntry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	err = d.Directory.Create(ctx, entry.ID, b, entry.Name)
	if err != nil {
		if errors.Is(err, jetstream.ErrInvalidKey) {
			return ErrNameInvalid
		}
		return err
	}

	return nil
}

func (d *AgentDirectory) Get(ctx context.Context, key string) (AgentDirectoryEntry, error) {
	b, err := d.Directory.Get(ctx, key)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

	var entry AgentDirectoryEntry
	err = json.Unmarshal(b, &entry)
	if err != nil {
		return AgentDirectoryEntry{}, err
	}

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
		err := json.Unmarshal(b, &entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

type AgentDirectoryEntry struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	GatewayID string           `json:"gateway_id"`
	Config    AgentBuildConfig `json:"config"`
}

type AgentBuildConfig struct {
	OS                string            `json:"os"`
	Arch              string            `json:"arch"`
	ServerURL         string            `json:"server_url"`
	ServerCertificate []byte            `json:"server_certificate"`
	Extra             map[string]string `json:"extra"`
}

type AgentHostDirectory struct {
	Directory
}

func (d *AgentHostDirectory) Create(ctx context.Context, entry AgentHostDirectoryEntry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	err = d.Directory.Create(ctx, entry.AgentID, b)
	if err != nil {
		if errors.Is(err, jetstream.ErrInvalidKey) {
			return ErrNameInvalid
		}
		return err
	}

	return nil
}

func (d *AgentHostDirectory) Get(ctx context.Context, key string) (AgentHostDirectoryEntry, error) {
	b, err := d.Directory.Get(ctx, key)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

	var entry AgentHostDirectoryEntry
	err = json.Unmarshal(b, &entry)
	if err != nil {
		return AgentHostDirectoryEntry{}, err
	}

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
		err := json.Unmarshal(b, &entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

type AgentHostDirectoryEntry struct {
	AgentID    string    `json:"agent_id"`
	Username   string    `json:"username"`
	Hostname   string    `json:"hostname"`
	System     string    `json:"system"`
	Address    string    `json:"address"`
	LastUpdate time.Time `json:"last_update"`
}

type GatewayDirectory struct {
	Directory
}

func (d *GatewayDirectory) Create(ctx context.Context, entry GatewayDirectoryEntry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	err = d.Directory.Create(ctx, entry.ID, b, entry.Name)
	if err != nil {
		if errors.Is(err, jetstream.ErrInvalidKey) {
			return ErrNameInvalid
		}
		return err
	}

	return nil
}

func (d *GatewayDirectory) Get(ctx context.Context, key string) (GatewayDirectoryEntry, error) {
	b, err := d.Directory.Get(ctx, key)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

	var entry GatewayDirectoryEntry
	err = json.Unmarshal(b, &entry)
	if err != nil {
		return GatewayDirectoryEntry{}, err
	}

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
		err := json.Unmarshal(b, &entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

type GatewayDirectoryEntry struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Config      GatewayConfig `json:"config"`
}

type GatewayConfig struct {
	URL         string            `json:"url"`
	Certificate []byte            `json:"certificate"`
	UserJWT     string            `json:"user_jwt"`
	UserKey     string            `json:"user_key"`
	Extra       map[string]string `json:"extra"`
}

type JobDirectory struct {
	Directory
}

func (d *JobDirectory) Create(ctx context.Context, entry JobDirectoryEntry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return d.Directory.Create(ctx, formatKey(entry.AgentID, entry.ID), b, entry.ID)
}

func (d *JobDirectory) Get(ctx context.Context, key string) (JobDirectoryEntry, error) {
	b, err := d.Directory.Get(ctx, key)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	var entry JobDirectoryEntry
	err = json.Unmarshal(b, &entry)
	if err != nil {
		return JobDirectoryEntry{}, err
	}

	return entry, nil
}

func (d *JobDirectory) List(ctx context.Context) ([]JobDirectoryEntry, error) {
	return d.list(ctx, formatKey("*", "*"))
}

func (d *JobDirectory) ListByAgentID(ctx context.Context, agentID string) ([]JobDirectoryEntry, error) {
	return d.list(ctx, formatKey(agentID, "*"))
}

func (d *JobDirectory) list(ctx context.Context, key string) ([]JobDirectoryEntry, error) {
	blobs, err := d.Directory.List(ctx, key)
	if err != nil {
		return nil, err
	}

	entries := make([]JobDirectoryEntry, 0, len(blobs))
	for _, b := range blobs {
		var entry JobDirectoryEntry
		err := json.Unmarshal(b, &entry)
		if err != nil {
			return nil, err
		}
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
