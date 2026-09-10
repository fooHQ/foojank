package directory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	UserKindAgent   = "agent"
	UserKindGateway = "gateway"
	UserKindClient  = "client"
)

type UserDirectory struct {
	Directory
}

func OpenUserDirectory(ctx context.Context, js jetstream.JetStream) (*UserDirectory, error) {
	dir, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: userDirectoryName,
	})
	if err != nil && !errors.Is(err, jetstream.ErrBucketExists) {
		return nil, err
	}
	return &UserDirectory{
		Directory: Directory{
			store: dir,
		},
	}, nil
}

func (d *UserDirectory) Create(ctx context.Context, entry UserDirectoryEntry) (UserDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return UserDirectoryEntry{}, err
	}

	rev, err := d.Directory.Create(ctx, entry.ID, b, entry.Name)
	if err != nil {
		return UserDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *UserDirectory) Update(ctx context.Context, entry UserDirectoryEntry) (UserDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return UserDirectoryEntry{}, err
	}

	rev, err := d.Directory.Update(ctx, entry.ID, b, entry.Revision)
	if err != nil {
		return UserDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *UserDirectory) Get(ctx context.Context, key string) (UserDirectoryEntry, error) {
	v, err := d.Directory.Get(ctx, key)
	if err != nil {
		return UserDirectoryEntry{}, err
	}

	var entry UserDirectoryEntry
	err = json.Unmarshal(v.Value, &entry)
	if err != nil {
		return UserDirectoryEntry{}, err
	}

	entry.Revision = v.Revision

	return entry, nil
}

func (d *UserDirectory) List(ctx context.Context) ([]UserDirectoryEntry, error) {
	blobs, err := d.Directory.List(ctx, formatKey("*"))
	if err != nil {
		return nil, err
	}

	entries := make([]UserDirectoryEntry, 0, len(blobs))
	for _, b := range blobs {
		var entry UserDirectoryEntry
		err := json.Unmarshal(b.Value, &entry)
		if err != nil {
			return nil, err
		}

		entry.Revision = b.Revision
		entries = append(entries, entry)
	}

	return entries, nil
}

func (d *UserDirectory) Delete(ctx context.Context, user UserDirectoryEntry) error {
	return d.Directory.Delete(ctx, user.ID, user.Name)
}

type UserDirectoryEntry struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Kind        string    `json:"kind"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Revision    uint64    `json:"-"`
}
