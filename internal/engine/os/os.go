package os

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	risoros "github.com/risor-io/risor/os"
)

var (
	ErrHandlerNotFound      = errors.New("handler not found")
	ErrCrossingFSBoundaries = errors.New("crossing filesystem boundaries")
)

var _ risoros.OS = &OS{}

type Option func(*OS)

func WithEnvVar(name, value string) Option {
	return func(o *OS) {
		o.environ[name] = value
	}
}

func WithArgs(args []string) Option {
	return func(o *OS) {
		o.args = args
	}
}

func WithStdin(file risoros.File) Option {
	return func(o *OS) {
		o.stdin = file
	}
}
func WithStdout(file risoros.File) Option {
	return func(o *OS) {
		o.stdout = file
	}
}

func WithURLHandler(scheme string, handler risoros.FS) Option {
	return func(o *OS) {
		o.urlHandlers[scheme] = handler
	}
}

func WithWorkDir(dir string) Option {
	return func(o *OS) {
		u, _ := toURL(dir)
		o.wd = u
	}
}

type ExitHandler func(int)

func WithExitHandler(handler ExitHandler) Option {
	return func(o *OS) {
		o.exitHandler = handler
	}
}

type OS struct {
	wd          *url.URL
	environ     map[string]string
	stdin       risoros.File
	stdout      risoros.File
	args        []string
	urlHandlers map[string]risoros.FS
	exitHandler ExitHandler
}

func (o *OS) Create(name string) (risoros.File, error) {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.Create(fURL.Path)
}

func (o *OS) Mkdir(name string, perm os.FileMode) error {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.Mkdir(fURL.Path, perm)
}

func (o *OS) MkdirAll(path string, perm os.FileMode) error {
	handler, fURL, ok := o.getRegisteredURLHandler(path)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.MkdirAll(fURL.Path, perm)
}

func (o *OS) MkdirTemp(dir, pattern string) (string, error) {
	// TODO
	_, _, ok := o.getRegisteredURLHandler(dir)
	if ok {
		return "", errors.New("creating temporary directory is not supported")
	}
	pth := o.joinWorkDir(dir)
	return os.MkdirTemp(pth, pattern)
}

func (o *OS) Open(name string) (risoros.File, error) {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.Open(fURL.Path)
}

func (o *OS) OpenFile(name string, flag int, perm os.FileMode) (risoros.File, error) {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.OpenFile(fURL.Path, flag, perm)
}

func (o *OS) ReadFile(name string) ([]byte, error) {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.ReadFile(fURL.Path)
}

func (o *OS) Remove(name string) error {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.Remove(fURL.Path)
}

func (o *OS) RemoveAll(path string) error {
	handler, fURL, ok := o.getRegisteredURLHandler(path)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.RemoveAll(fURL.Path)
}

func (o *OS) Rename(oldpath, newpath string) error {
	oldHandler, oldURL, ok := o.getRegisteredURLHandler(oldpath)
	if !ok {
		return ErrHandlerNotFound
	}
	newHandler, newURL, ok := o.getRegisteredURLHandler(newpath)
	if !ok {
		return ErrHandlerNotFound
	}
	if oldHandler != newHandler {
		return ErrCrossingFSBoundaries
	}
	return oldHandler.Rename(oldURL.Path, newURL.Path)
}

func (o *OS) Stat(name string) (os.FileInfo, error) {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.Stat(fURL.Path)
}

func (o *OS) Symlink(oldname, newname string) error {
	oldHandler, oldURL, ok := o.getRegisteredURLHandler(oldname)
	if !ok {
		return ErrHandlerNotFound
	}
	newHandler, newURL, ok := o.getRegisteredURLHandler(newname)
	if !ok {
		return ErrHandlerNotFound
	}
	if oldHandler != newHandler {
		return ErrCrossingFSBoundaries
	}
	return oldHandler.Symlink(oldURL.Path, newURL.Path)
}

func (o *OS) TempDir() string {
	return os.TempDir()
}

func (o *OS) WriteFile(name string, content []byte, perm os.FileMode) error {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.WriteFile(fURL.Path, content, perm)
}

func (o *OS) ReadDir(name string) ([]risoros.DirEntry, error) {
	handler, fURL, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}

	results, err := handler.ReadDir(fURL.Path)
	if err != nil {
		return nil, err
	}

	entries := make([]risoros.DirEntry, 0, len(results))
	for _, result := range results {
		entries = append(entries, &risoros.DirEntryWrapper{
			DirEntry: result,
		})
	}

	return entries, nil
}

func (o *OS) WalkDir(root string, fn risoros.WalkDirFunc) error {
	handler, fURL, ok := o.getRegisteredURLHandler(root)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.WalkDir(fURL.Path, fn)
}

func (o *OS) PathSeparator() rune {
	return os.PathSeparator
}

func (o *OS) PathListSeparator() rune {
	return os.PathListSeparator
}

func (o *OS) Chdir(dir string) error {
	handler, fURL, ok := o.getRegisteredURLHandler(dir)
	if !ok {
		return ErrHandlerNotFound
	}

	f, err := handler.Open(fURL.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("chdir " + fURL.Path + ": file is not a directory")
	}

	o.wd = fURL
	return nil
}

func (o *OS) Getwd() (dir string, err error) {
	wd := o.wd.Path
	if o.wd.Scheme != "file" {
		wd = o.wd.Scheme + "://" + o.wd.Host + wd
	}
	return wd, nil
}

func (o *OS) Stdout() risoros.File {
	return o.stdout
}

func (o *OS) Stdin() risoros.File {
	return o.stdin
}

func (o *OS) Args() []string {
	return o.args
}

func (o *OS) Environ() []string {
	var environ []string
	for k, v := range o.environ {
		environ = append(environ, k+"="+v)
	}
	return environ
}

func (o *OS) Getenv(key string) string {
	v, ok := o.environ[key]
	if !ok {
		return ""
	}
	return v
}

func (o *OS) Setenv(key, value string) error {
	o.environ[key] = value
	return nil
}

func (o *OS) Unsetenv(key string) error {
	delete(o.environ, key)
	return nil
}

func (o *OS) LookupEnv(key string) (string, bool) {
	v, ok := o.environ[key]
	return v, ok
}

func (o *OS) Exit(code int) {
	if o.exitHandler != nil {
		o.exitHandler(code)
	}
}

func (o *OS) Getpid() int {
	return os.Getpid()
}

func (o *OS) Getuid() int {
	return os.Getuid()
}

func (o *OS) Hostname() (string, error) {
	return os.Hostname()
}

func (o *OS) UserCacheDir() (string, error) {
	return os.UserCacheDir()
}

func (o *OS) UserConfigDir() (string, error) {
	return os.UserConfigDir()
}

func (o *OS) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (o *OS) joinWorkDir(name string) string {
	if !filepath.IsAbs(name) {
		return o.wd.JoinPath(name).Path
	}
	return filepath.Clean(name)
}

func (o *OS) getRegisteredURLHandler(path string) (risoros.FS, *url.URL, bool) {
	u, err := toURL(path)
	if err != nil {
		return nil, nil, false
	}

	u.Path = o.joinWorkDir(u.Path)
	handler, ok := o.urlHandlers[u.Scheme]
	return handler, u, ok
}

func toURL(path string) (*url.URL, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "file"
	}

	return u, nil
}

func NewContext(ctx context.Context, options ...Option) context.Context {
	o := &OS{
		wd:          initWD(),
		environ:     initEnviron(),
		urlHandlers: make(map[string]risoros.FS),
	}
	for _, option := range options {
		option(o)
	}

	return risoros.WithOS(ctx, o)
}

func initWD() *url.URL {
	wd, _ := os.Getwd()
	u, _ := toURL(wd)
	return u
}

func initEnviron() map[string]string {
	environ := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		environ[parts[0]] = parts[1]
	}
	return environ
}
