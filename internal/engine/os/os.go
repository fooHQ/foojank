package os

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path"
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

func WithURIHandler(scheme string, handler URIHandler) Option {
	return func(o *OS) {
		o.uriHandlers[scheme] = handler
	}
}

func WithWorkDir(dir string) Option {
	return func(o *OS) {
		u, _ := ToURL(dir)
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
	uriHandlers map[string]URIHandler
	exitHandler ExitHandler
}

func (o *OS) Create(name string) (risoros.File, error) {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return nil, err
	}

	pth := ToPath(fURL)
	return handler.Create(pth)
}

func (o *OS) Mkdir(name string, perm os.FileMode) error {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return err
	}

	pth := ToPath(fURL)
	return handler.Mkdir(pth, perm)
}

func (o *OS) MkdirAll(path string, perm os.FileMode) error {
	handler, fURL, err := o.getRegisteredURIHandler(path)
	if err != nil {
		return err
	}

	pth := ToPath(fURL)
	return handler.MkdirAll(pth, perm)
}

func (o *OS) MkdirTemp(dir, pattern string) (string, error) {
	return "", errors.New("not implemented")
}

func (o *OS) Open(name string) (risoros.File, error) {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return nil, err
	}

	pth := ToPath(fURL)
	return handler.Open(pth)
}

func (o *OS) OpenFile(name string, flag int, perm os.FileMode) (risoros.File, error) {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return nil, err
	}

	pth := ToPath(fURL)
	return handler.OpenFile(pth, flag, perm)
}

func (o *OS) ReadFile(name string) ([]byte, error) {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return nil, err
	}

	pth := ToPath(fURL)
	return handler.ReadFile(pth)
}

func (o *OS) Remove(name string) error {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return err
	}

	pth := ToPath(fURL)
	return handler.Remove(pth)
}

func (o *OS) RemoveAll(path string) error {
	handler, fURL, err := o.getRegisteredURIHandler(path)
	if err != nil {
		return err
	}

	pth := ToPath(fURL)
	return handler.RemoveAll(pth)
}

func (o *OS) Rename(oldpath, newpath string) error {
	oldHandler, oldURL, err := o.getRegisteredURIHandler(oldpath)
	if err != nil {
		return err
	}

	_, newURL, err := o.getRegisteredURIHandler(newpath)
	if err != nil {
		return err
	}

	if oldURL.Scheme != newURL.Scheme {
		return ErrCrossingFSBoundaries
	}

	oldPth := ToPath(oldURL)
	newPth := ToPath(newURL)
	return oldHandler.Rename(oldPth, newPth)
}

func (o *OS) Stat(name string) (os.FileInfo, error) {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return nil, err
	}

	pth := ToPath(fURL)
	return handler.Stat(pth)
}

func (o *OS) Symlink(oldname, newname string) error {
	oldHandler, oldURL, err := o.getRegisteredURIHandler(oldname)
	if err != nil {
		return err
	}

	_, newURL, err := o.getRegisteredURIHandler(newname)
	if err != nil {
		return err
	}

	if oldURL.Scheme != newURL.Scheme {
		return ErrCrossingFSBoundaries
	}

	oldPth := ToPath(oldURL)
	newPth := ToPath(newURL)
	return oldHandler.Symlink(oldPth, newPth)
}

func (o *OS) TempDir() string {
	return os.TempDir()
}

func (o *OS) WriteFile(name string, content []byte, perm os.FileMode) error {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return err
	}

	pth := ToPath(fURL)
	return handler.WriteFile(pth, content, perm)
}

func (o *OS) ReadDir(name string) ([]risoros.DirEntry, error) {
	handler, fURL, err := o.getRegisteredURIHandler(name)
	if err != nil {
		return nil, err
	}

	pth := ToPath(fURL)
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
	handler, fURL, err := o.getRegisteredURIHandler(root)
	if err != nil {
		return err
	}

	pth := ToPath(fURL)
	return handler.WalkDir(pth, fn)
}

func (o *OS) PathSeparator() rune {
	return os.PathSeparator
}

func (o *OS) PathListSeparator() rune {
	return os.PathListSeparator
}

func (o *OS) Chdir(dir string) error {
	handler, fURL, err := o.getRegisteredURIHandler(dir)
	if err != nil {
		return err
	}

	pth := ToPath(fURL)
	f, err := handler.Open(pth)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("chdir " + pth + ": file is not a directory")
	}

	o.wd = fURL
	return nil
}

func (o *OS) Getwd() (dir string, err error) {
	wd := o.wd.String()
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

func (o *OS) getRegisteredURIHandler(path string) (risoros.FS, *url.URL, error) {
	u, err := ToURL(path)
	if err != nil {
		return nil, nil, err
	}

	u = NormalizeURL(o.wd, u)
	handler, ok := o.uriHandlers[u.Scheme]
	if !ok {
		return nil, nil, ErrHandlerNotFound
	}

	fs, err := handler.GetFS(u)
	if err != nil {
		return nil, nil, err
	}

	return fs, u, nil
}

func NewContext(ctx context.Context, options ...Option) context.Context {
	o := &OS{
		wd:          initWD(),
		environ:     initEnviron(),
		uriHandlers: make(map[string]URIHandler),
	}
	for _, option := range options {
		option(o)
	}

	return risoros.WithOS(ctx, o)
}

func NormalizeURL(wd, u *url.URL) *url.URL {
	if u.Scheme == "" {
		if wd.Scheme == "" {
			u.Scheme = "file"
		} else {
			u.Scheme = wd.Scheme
		}
		u.Host = wd.Host
	}

	if !path.IsAbs(u.Path) {
		u.Path = path.Join(wd.Path, u.Path)
	} else {
		u.Path = path.Clean(u.Path)
	}

	return u
}

func initWD() *url.URL {
	wd, _ := os.Getwd()
	u, _ := ToURL(wd)
	// TODO: normalize URL!
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
