package nats

import (
	"context"
	"errors"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"

	memfs "github.com/foohq/foojank/filesystems/mem"
)

var (
	ErrDirectoriesNotSupported = errors.New("directories not supported")
	ErrSymlinksNotSupported    = errors.New("symlinks not supported")
	ErrBadDescriptor           = errors.New("bad descriptor")
	ErrIsDirectory             = memfs.ErrIsDirectory
)

type FS struct {
	cache   *memfs.FS
	store   jetstream.ObjectStore
	watcher jetstream.ObjectWatcher
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewFS(ctx context.Context, store jetstream.ObjectStore) (*FS, error) {
	// We probably do not need the cancellable context here since there yet no way to cancel it.
	watcher, err := store.Watch(ctx)
	if err != nil {
		return nil, err
	}

	cache, err := memfs.NewFS()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	fs := &FS{
		cache:   cache,
		store:   store,
		watcher: watcher,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Start watching for updates, which includes historical data
	go fs.watchUpdates()

	return fs, nil
}

// watchUpdates listens for NATS ObjectStore events to keep cache structure in sync
func (fs *FS) watchUpdates() {
	updateCh := fs.watcher.Updates()
	for {
		select {
		case update, ok := <-updateCh:
			if !ok {
				return
			}

			if update == nil { // Initial nil signals history complete
				continue
			}

			fs.mu.Lock()
			pth := cleanPath(update.Name)
			if update.Deleted {
				_ = fs.cache.RemoveAll(pth)
			} else if update.Opts != nil && update.Opts.Link != nil {
				// Handle symlinks
				_ = fs.cache.Symlink(update.Opts.Link.Name, pth)
			} else {
				// Create an empty file in cache to mark existence
				_ = fs.cache.MkdirAll(path.Dir(pth), 0755)
				_ = fs.cache.WriteFile(pth, nil, 0644)
			}
			fs.mu.Unlock()

		case <-fs.ctx.Done():
			return
		}
	}
}

// Create creates a new file (NATS + cache structure)
func (fs *FS) Create(name string) (risoros.File, error) {
	return fs.OpenFile(name, risoros.O_RDWR|risoros.O_CREATE|risoros.O_TRUNC, 0666)
}

// Mkdir returns an error as explicit directory creation is not supported
func (fs *FS) Mkdir(_ string, _ risoros.FileMode) error {
	return ErrDirectoriesNotSupported
}

// MkdirAll returns an error as explicit directory creation is not supported
func (fs *FS) MkdirAll(_ string, _ risoros.FileMode) error {
	return ErrDirectoriesNotSupported
}

// Open opens a file for reading (fetches content from NATS)
func (fs *FS) Open(name string) (risoros.File, error) {
	return fs.OpenFile(name, risoros.O_RDONLY, 0)
}

// OpenFile opens a file with specified flags (cache structure + NATS for content)
func (fs *FS) OpenFile(name string, flag int, perm risoros.FileMode) (risoros.File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	pth := cleanPath(name)

	// Check cache for existence
	_, err := fs.cache.Stat(pth)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	exists := err == nil
	if !exists && flag&risoros.O_CREATE == 0 {
		return nil, os.ErrNotExist
	}

	if flag&risoros.O_CREATE != 0 {
		// Create on NATS
		if _, err := fs.store.Put(fs.ctx, jetstream.ObjectMeta{Name: pth}, strings.NewReader("")); err != nil {
			return nil, err
		}

		if err := fs.cache.MkdirAll(path.Dir(pth), 0755); err != nil {
			return nil, err
		}

		// Update cache structure
		if err := fs.cache.WriteFile(pth, nil, perm); err != nil {
			return nil, err
		}
	}

	f, err := fs.cache.OpenFile(pth, flag, perm)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// Fetch content from NATS for reading
	var o jetstream.ObjectResult
	if flag&risoros.O_WRONLY == 0 && !info.IsDir() {
		o, err = fs.store.Get(fs.ctx, pth)
		if err != nil {
			return nil, err
		}
	}

	return &natsFile{
		File: f,
		fs:   fs,
		flag: flag,
		path: pth,
		obj:  o,
	}, nil
}

// ReadFile reads the entire contents of a file from NATS
func (fs *FS) ReadFile(name string) ([]byte, error) {
	f, err := fs.OpenFile(name, risoros.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	b := make([]byte, info.Size())
	_, err = f.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Remove removes a file
func (fs *FS) Remove(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	pth := cleanPath(name)
	err := fs.store.Delete(fs.ctx, pth)
	if err != nil {
		return err
	}

	return fs.cache.Remove(pth)
}

// RemoveAll removes a path and its implied children
func (fs *FS) RemoveAll(pth string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	objects, err := fs.store.List(fs.ctx)
	if err != nil && !errors.Is(err, jetstream.ErrNoObjectsFound) {
		return err
	}

	pth = cleanPath(pth)
	for _, obj := range objects {
		if strings.HasPrefix(obj.Name, pth) {
			if err := fs.store.Delete(fs.ctx, obj.Name); err != nil {
				return err
			}
		}
	}
	return fs.cache.RemoveAll(pth)
}

// Rename renames a file
func (fs *FS) Rename(oldpath, newpath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	oldpath = cleanPath(oldpath)
	newpath = cleanPath(newpath)
	err := fs.store.UpdateMeta(fs.ctx, oldpath, jetstream.ObjectMeta{Name: newpath})
	if err != nil {
		return err
	}

	return fs.cache.Rename(oldpath, newpath)
}

// Stat returns file information (cache only)
func (fs *FS) Stat(name string) (risoros.FileInfo, error) {
	return fs.cache.Stat(cleanPath(name))
}

// Symlink creates a symbolic link
func (fs *FS) Symlink(oldname, newname string) error {
	return ErrSymlinksNotSupported
}

// WriteFile writes data to a file
func (fs *FS) WriteFile(name string, data []byte, perm risoros.FileMode) error {
	f, err := fs.OpenFile(name, risoros.O_WRONLY|risoros.O_CREATE, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// ReadDir reads directory contents from the cache
func (fs *FS) ReadDir(name string) ([]risoros.DirEntry, error) {
	name = cleanPath(name)
	return fs.cache.ReadDir(name)
}

// WalkDir walks the directory tree in the cache
func (fs *FS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	return fs.cache.WalkDir(cleanPath(root), fn)
}

// Close shuts down the filesystem
func (fs *FS) Close() error {
	err := fs.watcher.Stop()
	if err != nil {
		return err
	}
	fs.cancel()
	return nil
}

// natsFile wraps a memfs.FS file to sync writes back to NATS
type natsFile struct {
	risoros.File
	fs      *FS
	flag    int
	path    string
	obj     jetstream.ObjectResult
	content []byte
	offset  int64
	mu      sync.Mutex
}

func (f *natsFile) Read(b []byte) (int, error) {
	if f.flag&risoros.O_WRONLY != 0 {
		return 0, ErrBadDescriptor
	}

	if f.obj == nil {
		return 0, os.ErrInvalid
	}

	return f.obj.Read(b)
}

func (f *natsFile) Stat() (risoros.FileInfo, error) {
	if f.obj != nil {
		info, err := f.obj.Info()
		if err != nil {
			return nil, err
		}
		return &fileInfo{
			info: info,
		}, nil
	}
	return f.fs.cache.Stat(f.path)
}

// Write overrides the underlying file write to sync to NATS
func (f *natsFile) Write(b []byte) (int, error) {
	if f.flag&risoros.O_WRONLY == 0 && f.flag&risoros.O_RDWR == 0 {
		return 0, ErrBadDescriptor
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.content = append(f.content[:f.offset], b...)
	f.offset += int64(len(b))
	return len(b), nil
}

// Close syncs any final changes to NATS (if needed)
func (f *natsFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.obj != nil {
		err := f.obj.Close()
		if err != nil {
			return err
		}
	}

	if len(f.content) > 0 {
		_, err := f.fs.store.PutBytes(f.fs.ctx, f.path, f.content)
		if err != nil {
			return err
		}
	}

	err := f.File.Close()
	if err != nil {
		return err
	}

	return nil
}

type fileInfo struct {
	info *jetstream.ObjectInfo
}

func (f *fileInfo) Name() string {
	return f.info.Name
}

func (f *fileInfo) Size() int64 {
	return int64(f.info.Size)
}

func (f *fileInfo) Mode() risoros.FileMode {
	return 0666
}

func (f *fileInfo) ModTime() time.Time {
	return f.info.ModTime
}

func (f *fileInfo) IsDir() bool {
	return false
}

func (f *fileInfo) Sys() any {
	return nil
}

// Helper function to clean paths
func cleanPath(pth string) string {
	return path.Clean("/" + pth)
}
