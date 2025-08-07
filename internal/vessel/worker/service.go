package worker

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/foohq/ren"
	"github.com/foohq/ren/modules"
	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/reader"
	"github.com/foohq/foojank/internal/vessel/worker/writer"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	ID            string
	Stream        string
	StdinSubject  string
	StdoutSubject string
	UpdateSubject string
	Entrypoint    string
	Args          []string
	Env           []string
	Connection    jetstream.JetStream
	Filesystems   map[string]risoros.FS
	EventCh       chan<- any
}

type Service struct {
	args Arguments
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	log.Debug("Service started", "service", "vessel.workmanager.worker")
	defer log.Debug("Service stopped", "service", "vessel.workmanager.worker")

	s.sendEvent(ctx, EventWorkerStarted{
		ID: s.args.ID,
	})
	// IMPORTANT: Send must not check context state lest the message will be lost.
	defer s.sendEvent(context.Background(), EventWorkerStopped{
		ID: s.args.ID,
	})

	stdin := ren.NewPipe()
	stdout := ren.NewPipe()

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return reader.New(reader.Arguments{
			Connection: s.args.Connection,
			Stream:     s.args.Stream,
			Subject:    s.args.StdinSubject,
			File:       stdin,
		}).Start(groupCtx)
	})

	group.Go(func() error {
		return writer.New(writer.Arguments{
			Connection: s.args.Connection,
			File:       stdout,
			Subject:    s.args.StdoutSubject,
		}).Start(groupCtx)
	})

	code, _ := run(groupCtx, s.args.Entrypoint, s.args.Args, s.args.Env, stdin, stdout, s.args.Filesystems)
	// TODO: pass error to UpdateJob
	b, err := proto.Marshal(proto.UpdateJob{
		JobID:      s.args.ID,
		ExitStatus: int64(code),
	})
	if err != nil {
		log.Debug("Cannot encode message", "error", err)
		return err
	}

	// Reload context with timeout if the parent context is done.
	// The newly created context is used to publish an update job message.
	// This is the best effort delivery attempt. Clients should be able
	// to handle the case when the message is not delivered.
	pubCtx, cancel := reloadContextWithTimeout(groupCtx, 4*time.Second)
	defer cancel()

	_, err = s.args.Connection.Publish(pubCtx, s.args.UpdateSubject, b)
	if err != nil {
		log.Debug("Cannot publish message", "error", err)
		return err
	}

	<-groupCtx.Done()
	_ = stdin.Close()
	_ = stdout.Close()

	return group.Wait()
}

func (s *Service) sendEvent(ctx context.Context, event any) {
	select {
	case s.args.EventCh <- event:
	case <-ctx.Done():
	}
}

const (
	exitFailure = 1
)

func run(ctx context.Context, entrypoint string, args, env []string, stdin, stdout risoros.File, filesystems map[string]risoros.FS) (int, error) {
	u, err := url.Parse(entrypoint)
	if err != nil {
		return exitFailure, err
	}

	fsType := u.Scheme
	if fsType == "" {
		fsType = "file"
	}

	targetFS, ok := filesystems[fsType]
	if !ok {
		return exitFailure, errors.New("filesystem not found")
	}

	b, err := targetFS.ReadFile(u.Path)
	if err != nil {
		return exitFailure, errors.New("cannot read package '" + u.Path + "': " + err.Error())
	}

	opts := []ren.Option{
		ren.WithArgs(args),
		ren.WithStdin(stdin),
		ren.WithStdout(stdout),
		ren.WithFilesystems(filesystems),
	}

	// Configure exit status handler
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var status int
	opts = append(opts, ren.WithExitHandler(func(code int) {
		log.Debug("on exit", "code", code)
		status = code
		cancel()
	}))

	// Configure modules
	for _, name := range modules.Modules() {
		mod, ok := modules.Module(name)
		if !ok {
			continue
		}
		opts = append(opts, ren.WithModule(mod))
	}

	// Configure environment variables
	for i := 0; i < len(env); i += 2 {
		name := env[i]
		value := ""
		if i+1 < len(env) {
			value = env[i+1]
		}
		opts = append(opts, ren.WithEnvVar(name, value))
	}

	err = ren.RunBytes(
		ctx,
		b,
		opts...,
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		return exitFailure, err
	}

	return status, nil
}

// reloadContextWithTimeout creates a new context with specified timeout if ctx is done, otherwise returns ctx.
func reloadContextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	newCtx := ctx
	cancel := func() {}
	select {
	case <-ctx.Done():
		newCtx, cancel = context.WithTimeout(context.Background(), timeout)
	default:
	}
	return newCtx, cancel
}

type (
	EventWorkerStarted struct {
		ID string
	}
	EventWorkerStopped struct {
		ID string
	}
)
