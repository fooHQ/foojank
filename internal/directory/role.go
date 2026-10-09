package directory

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type RoleDirectory struct {
	Directory
}

func OpenRoleDirectory(ctx context.Context, js jetstream.JetStream) (*RoleDirectory, error) {
	dir, err := Open(ctx, js, roleDirectoryName)
	if err != nil {
		return nil, err
	}
	return &RoleDirectory{
		Directory: dir,
	}, nil
}

func (d *RoleDirectory) Create(ctx context.Context, entry RoleDirectoryEntry) (RoleDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return RoleDirectoryEntry{}, err
	}

	rev, err := d.Directory.Create(ctx, entry.Name, b)
	if err != nil {
		return RoleDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *RoleDirectory) Update(ctx context.Context, entry RoleDirectoryEntry) (RoleDirectoryEntry, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return RoleDirectoryEntry{}, err
	}

	rev, err := d.Directory.Update(ctx, entry.Name, b, entry.Revision)
	if err != nil {
		return RoleDirectoryEntry{}, err
	}

	entry.Revision = rev
	return entry, nil
}

func (d *RoleDirectory) Get(ctx context.Context, key string) (RoleDirectoryEntry, error) {
	v, err := d.Directory.Get(ctx, key)
	if err != nil {
		return RoleDirectoryEntry{}, err
	}

	var entry RoleDirectoryEntry
	err = json.Unmarshal(v.Value, &entry)
	if err != nil {
		return RoleDirectoryEntry{}, err
	}

	entry.Revision = v.Revision

	return entry, nil
}

func (d *RoleDirectory) List(ctx context.Context) ([]RoleDirectoryEntry, error) {
	blobs, err := d.Directory.List(ctx, formatKey("*"))
	if err != nil {
		return nil, err
	}

	entries := make([]RoleDirectoryEntry, 0, len(blobs))
	for _, b := range blobs {
		var entry RoleDirectoryEntry
		err := json.Unmarshal(b.Value, &entry)
		if err != nil {
			return nil, err
		}

		entry.Revision = b.Revision
		entries = append(entries, entry)
	}

	return entries, nil
}

func (d *RoleDirectory) Delete(ctx context.Context, role RoleDirectoryEntry) error {
	return d.Directory.Delete(ctx, role.Name)
}

type RoleDirectoryEntry struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Privileges  []string  `json:"privileges"`
	CreatedAt   time.Time `json:"created_at"`
	Revision    uint64    `json:"-"`
}
