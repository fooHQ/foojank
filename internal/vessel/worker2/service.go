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
	"github.com/foohq/foojank/internal/vessel/worker2/reader"
	"github.com/foohq/foojank/internal/vessel/worker2/writer"
	"github.com/foohq/foojank/proto"
)

const (
	ExitFailure = 1
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
	EventCh       chan<- struct{}
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
	stdin := ren.NewPipe()
	stdout := ren.NewPipe()

	runnerOutCh := make(chan any)
	encoderOutCh := make(chan []byte)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return runner(groupCtx, s.args.ID, s.args.Entrypoint, s.args.Args, s.args.Env, stdin, stdout, s.args.Filesystems, runnerOutCh)
	})

	group.Go(func() error {
		return encoder(groupCtx, runnerOutCh, encoderOutCh)
	})

	group.Go(func() error {
		return publisher(groupCtx, s.args.Connection, s.args.UpdateSubject, encoderOutCh)
	})

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

	<-groupCtx.Done()
	_ = stdin.Close()
	_ = stdout.Close()

	err := group.Wait()
	if err != nil {
		log.Debug("worker stopped", "error", err)
		return err
	}

	return nil
}

func runner(ctx context.Context, id, entrypoint string, args, env []string, stdin, stdout risoros.File, filesystems map[string]risoros.FS, outputCh chan any) error {
	code, _ := run(ctx, entrypoint, args, env, stdin, stdout, filesystems)
	// TODO: pass error to UpdateJob except context.Canceled

	// TODO: figure out how to send UpdateJob when context is canceled

	select {
	case outputCh <- proto.UpdateJob{
		JobID:      id,
		ExitStatus: int64(code),
	}:
	case <-ctx.Done():
		return nil
	}

	return nil
}

func encoder(ctx context.Context, inputCh <-chan any, outputCh chan<- []byte) error {
	select {
	case msg := <-inputCh:
		b, err := proto.Marshal(msg)
		if err != nil {
			log.Debug("cannot encode message", "error", err)
			return err
		}

		select {
		case outputCh <- b:
		case <-ctx.Done():
			return nil
		}
		return nil

	case <-ctx.Done():
		return nil
	}
}

func publisher(ctx context.Context, conn jetstream.JetStream, subject string, inputCh <-chan []byte) error {
	select {
	case msg := <-inputCh:
		_, err := conn.Publish(ctx, subject, msg)
		if err != nil {
			log.Debug("cannot publish message", "error", err)
			return err
		}
		// Trigger error to start group termination
		return errors.New("terminated")

	case <-ctx.Done():
		return nil
	}
}

func run(ctx context.Context, entrypoint string, args, env []string, stdin, stdout risoros.File, filesystems map[string]risoros.FS) (int, error) {
	u, err := url.Parse(entrypoint)
	if err != nil {
		return ExitFailure, err
	}

	fsType := u.Scheme
	if fsType == "" {
		fsType = "file"
	}

	targetFS, ok := filesystems[fsType]
	if !ok {
		return ExitFailure, errors.New("filesystem not found")
	}

	b, err := targetFS.ReadFile(u.Path)
	if err != nil {
		return ExitFailure, errors.New("cannot read package '" + u.Path + "': " + err.Error())
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
		return ExitFailure, err
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
