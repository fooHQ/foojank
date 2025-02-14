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
	ErrHandlerNotFound = errors.New("handler not found")
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

func WithURLHandler(scheme, host string, handler risoros.FS) Option {
	return func(o *OS) {
		key := strings.Join([]string{scheme, host}, "/")
		o.urlHandlers[key] = handler
	}
}

func WithWorkDir(dir string) Option {
	return func(o *OS) {
		o.wd = filepath.Join(string(os.PathSeparator), dir)
	}
}

type ExitHandler func(int)

func WithExitHandler(handler ExitHandler) Option {
	return func(o *OS) {
		o.exitHandler = handler
	}
}

type OS struct {
	wd          string
	environ     map[string]string
	stdin       risoros.File
	stdout      risoros.File
	args        []string
	urlHandlers map[string]risoros.FS
	exitHandler ExitHandler
}

// TODO: joint with wd if file://

func (o *OS) Create(name string) (risoros.File, error) {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	// TODO: joint with wd if file://
	return handler.Create(pth)
}

func (o *OS) Mkdir(name string, perm os.FileMode) error {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return ErrHandlerNotFound
	}
	// TODO: joint with wd if file://
	return handler.Mkdir(pth, perm)
}

func (o *OS) MkdirAll(path string, perm os.FileMode) error {
	handler, pth, ok := o.getRegisteredURLHandler(path)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.MkdirAll(pth, perm)
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
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.Open(pth)
}

func (o *OS) OpenFile(name string, flag int, perm os.FileMode) (risoros.File, error) {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.OpenFile(pth, flag, perm)
}

func (o *OS) ReadFile(name string) ([]byte, error) {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.ReadFile(pth)
}

func (o *OS) Remove(name string) error {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.Remove(pth)
}

func (o *OS) RemoveAll(path string) error {
	handler, pth, ok := o.getRegisteredURLHandler(path)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.RemoveAll(pth)
}

func (o *OS) Rename(oldpath, newpath string) error {
	oldHandler, oldPth, ok := o.getRegisteredURLHandler(oldpath)
	if !ok {
		return ErrHandlerNotFound
	}
	newHandler, newPth, ok := o.getRegisteredURLHandler(newpath)
	if !ok {
		return ErrHandlerNotFound
	}
	if oldHandler != newHandler {
		return errors.New("rename cannot pass filesystem boundaries")
	}
	return oldHandler.Rename(oldPth, newPth)
}

func (o *OS) Stat(name string) (os.FileInfo, error) {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return handler.Stat(pth)
}

func (o *OS) Symlink(oldname, newname string) error {
	oldHandler, oldPth, ok := o.getRegisteredURLHandler(oldname)
	if !ok {
		return ErrHandlerNotFound
	}
	newHandler, newPth, ok := o.getRegisteredURLHandler(newname)
	if !ok {
		return ErrHandlerNotFound
	}
	if oldHandler != newHandler {
		return errors.New("symlink cannot pass filesystem boundaries")
	}
	return oldHandler.Symlink(oldPth, newPth)
}

func (o *OS) TempDir() string {
	return os.TempDir()
}

func (o *OS) WriteFile(name string, content []byte, perm os.FileMode) error {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.WriteFile(pth, content, perm)
}

func (o *OS) ReadDir(name string) ([]risoros.DirEntry, error) {
	handler, pth, ok := o.getRegisteredURLHandler(name)
	if !ok {
		return nil, ErrHandlerNotFound
	}

	results, err := handler.ReadDir(pth)
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
	handler, pth, ok := o.getRegisteredURLHandler(root)
	if !ok {
		return ErrHandlerNotFound
	}
	return handler.WalkDir(pth, fn)
}

func (o *OS) PathSeparator() rune {
	return os.PathSeparator
}

func (o *OS) PathListSeparator() rune {
	return os.PathListSeparator
}

func (o *OS) Chdir(dir string) error {
	_, _, ok := o.getRegisteredURLHandler(dir)
	if ok {
		return errors.New("chdir is not supported")
	}

	pth := o.joinWorkDir(dir)
	f, err := os.Open(pth)
	if err != nil {
		return err
	}
	defer f.Close()

	// Checks whether the file is a directory by trying to read the entries.
	_, err = f.Readdirnames(0)
	if err != nil {
		// Trying hard to return the same error string as stdlib's os.Chdir.
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			return errors.New("chdir " + pth + ": " + pathErr.Unwrap().Error())
		}
		return err
	}

	o.wd = pth
	return nil
}

func (o *OS) Getwd() (dir string, err error) {
	return o.wd, nil
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
		return filepath.Join(o.wd, name)
	}
	return name
}

func (o *OS) getRegisteredURLHandler(path string) (risoros.FS, string, bool) {
	// TODO: support windows paths!
	u, err := url.Parse(path)
	if err != nil {
		return nil, "", false
	}

	if u.Scheme == "" {
		u.Scheme = "file"
	}

	if u.Scheme == "file" {
		u.Path = o.joinWorkDir(u.Path)
	}

	key := strings.Join([]string{u.Scheme, u.Host}, "/")
	handler, ok := o.urlHandlers[key]
	return handler, u.Path, ok
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

func initWD() string {
	wd, _ := os.Getwd()
	return wd
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
