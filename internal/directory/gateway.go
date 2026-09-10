package directory

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
)

type GatewayDirectory struct {
	Directory
}

func OpenGatewayDirectory(ctx context.Context, js jetstream.JetStream) (*GatewayDirectory, error) {
	dir, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: gatewayDirectoryName,
	})
	if err != nil && !errors.Is(err, jetstream.ErrBucketExists) {
		return nil, err
	}
	return &GatewayDirectory{
		Directory: Directory{
			store: dir,
		},
	}, nil
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
