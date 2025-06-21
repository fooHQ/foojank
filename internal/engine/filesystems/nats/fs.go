package nats

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"

	memfs "github.com/foohq/foojank/internal/engine/filesystems/mem"
)

var (
	ErrInvalid              = os.ErrInvalid
	ErrNotExist             = os.ErrNotExist
	ErrIsDirectory          = memfs.ErrIsDirectory
	ErrSymlinksNotSupported = errors.New("symlinks not supported")
	ErrBadDescriptor        = errors.New("bad descriptor")
	ErrNotSynced            = errors.New("filesystem not synchronized")
)

type FS struct {
	cache   *memfs.FS
	store   jetstream.ObjectStore
	watcher jetstream.ObjectWatcher
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	synced  chan struct{}
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
		synced:  make(chan struct{}),
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
				close(fs.synced)
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
				if isDir(update) {
					_ = fs.cache.MkdirAll(pth, 0755)
				} else if isFile(update) {
					// Create an empty file in cache to mark existence
					_ = fs.cache.MkdirAll(path.Dir(pth), 0755)
					_ = fs.cache.WriteFile(pth, nil, 0644)
				}
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
func (fs *FS) Mkdir(name string, perm risoros.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.isSynced() {
		return ErrNotSynced
	}

	pth := cleanPath(name)

	revert, err := fs.mkdirCache(pth, perm)
	if err != nil {
		return err
	}

	if _, err := fs.store.Put(fs.ctx, newObjectMetadata(pth, typeDir), strings.NewReader("")); err != nil {
		revert()
		return &Error{err}
	}

	return nil
}

// MkdirAll returns an error as explicit directory creation is not supported
func (fs *FS) MkdirAll(pth string, perm risoros.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.isSynced() {
		return ErrNotSynced
	}

	pth = cleanPath(pth)

	revert, err := fs.mkdirAllCache(pth, perm)
	if err != nil {
		return err
	}

	if _, err := fs.store.Put(fs.ctx, newObjectMetadata(pth, typeDir), strings.NewReader("")); err != nil {
		revert()
		return &Error{err}
	}

	return nil
}

func (fs *FS) mkdirCache(pth string, perm risoros.FileMode) (func(), error) {
	err := fs.cache.Mkdir(pth, perm)
	if err != nil {
		return nil, err
	}

	revertFn := func() {
		_ = fs.cache.Remove(pth)
	}

	return revertFn, nil
}

func (fs *FS) mkdirAllCache(pth string, perm risoros.FileMode) (func(), error) {
	var createdDirs []string
	var parentDir string

	for _, dir := range strings.Split(pth, "/") {
		dirPth := cleanPath(path.Join(parentDir, dir))
		parentDir = dirPth
		err := fs.cache.Mkdir(dirPth, perm)
		if err != nil {
			if !errors.Is(err, memfs.ErrExist) {
				return nil, err
			}
			continue
		}

		createdDirs = append(createdDirs, dirPth)
	}

	revertFn := func() {
		for i := len(createdDirs) - 1; i >= 0; i-- {
			_ = fs.cache.Remove(createdDirs[i])
		}
	}

	return revertFn, nil
}

func (fs *FS) writeFileCache(name string, data []byte, perm risoros.FileMode) (func(), error) {
	if err := fs.cache.WriteFile(name, data, perm); err != nil {
		return nil, err
	}

	revertFn := func() {
		_ = fs.cache.Remove(name)
	}

	return revertFn, nil
}

// Open opens a file for reading (fetches content from NATS)
func (fs *FS) Open(name string) (risoros.File, error) {
	return fs.OpenFile(name, risoros.O_RDONLY, 0)
}

// OpenFile opens a file with specified flags (cache structure + NATS for content)
func (fs *FS) OpenFile(name string, flag int, perm risoros.FileMode) (risoros.File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.isSynced() {
		return nil, ErrNotSynced
	}

	pth := cleanPath(name)

	// Check cache for existence
	_, err := fs.cache.Stat(pth)
	if err != nil && !errors.Is(err, memfs.ErrNotExist) {
		return nil, err
	}

	exists := err == nil
	if !exists && flag&risoros.O_CREATE == 0 {
		return nil, ErrNotExist
	}

	if flag&risoros.O_CREATE != 0 {
		revertDirs, err := fs.mkdirCache(path.Dir(pth), 0755)
		if err != nil && !errors.Is(err, memfs.ErrExist) {
			return nil, err
		}

		revertFile, err := fs.writeFileCache(pth, nil, perm)
		if err != nil {
			revertDirs()
			return nil, err
		}

		// Create on NATS
		if _, err := fs.store.Put(fs.ctx, newObjectMetadata(pth, typeFile), strings.NewReader("")); err != nil {
			revertFile()
			revertDirs()
			return nil, &Error{err}
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
	if flag&risoros.O_WRONLY == 0 {
		o, err = fs.store.Get(fs.ctx, pth)
		if err != nil {
			if !errors.Is(err, jetstream.ErrObjectNotFound) || !info.IsDir() {
				return nil, &Error{err}
			}
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

	if !fs.isSynced() {
		return ErrNotSynced
	}

	pth := cleanPath(name)
	err := fs.store.Delete(fs.ctx, pth)
	if err != nil {
		return &Error{err}
	}

	err = fs.cache.Remove(pth)
	if err != nil {
		return err
	}

	return nil
}

// RemoveAll removes a path and its implied children
func (fs *FS) RemoveAll(pth string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.isSynced() {
		return ErrNotSynced
	}

	objects, err := fs.store.List(fs.ctx)
	if err != nil && !errors.Is(err, jetstream.ErrNoObjectsFound) {
		return &Error{err}
	}

	pth = cleanPath(pth)
	for _, obj := range objects {
		if strings.HasPrefix(obj.Name, pth) {
			if err := fs.store.Delete(fs.ctx, obj.Name); err != nil {
				return &Error{err}
			}
		}
	}

	return fs.cache.RemoveAll(pth)
}

// Rename renames a file
func (fs *FS) Rename(oldpath, newpath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.isSynced() {
		return ErrNotSynced
	}

	oldpath = cleanPath(oldpath)
	newpath = cleanPath(newpath)
	err := fs.store.UpdateMeta(fs.ctx, oldpath, jetstream.ObjectMeta{Name: newpath})
	if err != nil {
		return &Error{err}
	}

	return fs.cache.Rename(oldpath, newpath)
}

// Stat returns file information (cache only)
func (fs *FS) Stat(name string) (risoros.FileInfo, error) {
	if !fs.isSynced() {
		return nil, ErrNotSynced
	}

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
	if !fs.isSynced() {
		return nil, ErrNotSynced
	}

	name = cleanPath(name)
	return fs.cache.ReadDir(name)
}

// WalkDir walks the directory tree in the cache
func (fs *FS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	if !fs.isSynced() {
		return ErrNotSynced
	}

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

func (fs *FS) Wait(ctx context.Context) error {
	select {
	case <-fs.synced:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (fs *FS) isSynced() bool {
	select {
	case <-fs.synced:
		return true
	default:
		return false
	}
}

var (
	_ io.ReaderFrom = &natsFile{}
	_ io.WriterTo   = &natsFile{}
)

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
		return 0, ErrInvalid
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

func (f *natsFile) ReadFrom(r io.Reader) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	info, err := f.fs.store.Put(f.fs.ctx, newObjectMetadata(f.path, typeFile), r)
	if err != nil {
		return 0, &Error{err}
	}

	return int64(info.Size), nil
}

func (f *natsFile) WriteTo(w io.Writer) (n int64, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.obj == nil {
		return 0, ErrInvalid
	}

	var total int
	b := make([]byte, 2048)
	for {
		n, err := f.obj.Read(b)
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}

		// Is EOF?
		if n == 0 {
			break
		}

		total += n

		_, err = w.Write(b[:n])
		if err != nil {
			return 0, err
		}
	}

	return int64(total), nil
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
		_, err := f.fs.store.Put(f.fs.ctx, newObjectMetadata(f.path, typeFile), bytes.NewReader(f.content))
		if err != nil {
			return &Error{err}
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
	return isDir(f.info)
}

func (f *fileInfo) Sys() any {
	return nil
}

// Helper function to clean paths
func cleanPath(pth string) string {
	return path.Clean("/" + pth)
}

const (
	metadataKeyType = "file-type"
)

const (
	typeFile = "f"
	typeDir  = "d"
)

func newObjectMetadata(name string, fileType string) jetstream.ObjectMeta {
	return jetstream.ObjectMeta{
		Name: name,
		Metadata: map[string]string{
			metadataKeyType: fileType,
		},
	}
}

func isFile(info *jetstream.ObjectInfo) bool {
	v, ok := info.Metadata[metadataKeyType]
	return ok && v == typeFile
}

func isDir(info *jetstream.ObjectInfo) bool {
	v, ok := info.Metadata[metadataKeyType]
	return ok && v == typeDir
}

type Error struct {
	err error
}

func (e *Error) Error() string {
	switch {
	case errors.Is(e.err, jetstream.ErrBucketExists):
		return "repository already exists"
	case errors.Is(e.err, jetstream.ErrBucketNotFound):
		return "repository not found"
	case errors.Is(e.err, jetstream.ErrObjectNotFound):
		return "file not found"
	case errors.Is(e.err, jetstream.ErrInvalidStoreName):
		return "invalid repository name"
	}
	return e.err.Error()
}
