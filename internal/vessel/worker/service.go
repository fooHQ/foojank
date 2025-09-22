package worker

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/foohq/ren"
	"github.com/foohq/ren/modules"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/message"
)

var errDone = errors.New("done")

type Arguments struct {
	ID          string
	File        string
	Args        []string
	Env         []string
	EventCh     chan<- message.Msg
	StdinCh     <-chan []byte
	Filesystems map[string]risoros.FS
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
	log.Debug("Service started", "service", "vessel.workmanager.worker", "id", s.args.ID)
	defer log.Debug("Service stopped", "service", "vessel.workmanager.worker", "id", s.args.ID)

	// IMPORTANT: Send event must not check the current context lest the message will be lost.
	s.sendEvent(context.Background(), EventWorkerStarted{
		WorkerID: s.args.ID,
	})

	stdin := ren.NewPipe()
	stdout := ren.NewPipe()

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return stdinWriter(groupCtx, s.args.StdinCh, stdin)
	})

	group.Go(func() error {
		return stdoutReader(groupCtx, s.args.ID, stdout, s.args.EventCh)
	})

	group.Go(func() error {
		code, err := run(groupCtx, s.args.File, s.args.Args, s.args.Env, stdin, stdout, s.args.Filesystems)
		if err != nil {
			log.Debug("Run failed", "error", err)
		}

		// IMPORTANT: Send must not check context state lest the message will be lost.
		s.sendEvent(context.Background(), EventWorkerStopped{
			WorkerID: s.args.ID,
			Status:   code,
			Error:    err,
		})

		// The error will trigger a group shutdown which will lead to worker shutdown.
		return errDone
	})

	<-groupCtx.Done()
	_ = stdin.Close()
	_ = stdout.Close()

	err := group.Wait()
	if err != nil && !errors.Is(err, errDone) {
		return err
	}

	return nil
}

func stdinWriter(ctx context.Context, inputCh <-chan []byte, outputFile risoros.File) error {
	log.Debug("Service started", "service", "vessel.worker.stdinwriter")
	defer log.Debug("Service stopped", "service", "vessel.worker.stdinwriter")

	for {
		select {
		case data := <-inputCh:
			_, err := outputFile.Write(data)
			if err != nil {
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func stdoutReader(ctx context.Context, workerID string, inputFile risoros.File, outputCh chan<- message.Msg) error {
	log.Debug("Service started", "service", "vessel.worker.stdoutwriter")
	defer log.Debug("Service stopped", "service", "vessel.worker.stdoutwriter")

	for {
		b := make([]byte, 4096)
		n, err := inputFile.Read(b)
		if err != nil {
			return nil
		}

		select {
		case outputCh <- EventWorkerOutput{
			WorkerID:   workerID,
			OutputData: b[:n],
		}:
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) sendEvent(ctx context.Context, event message.Msg) {
	select {
	case s.args.EventCh <- event:
	case <-ctx.Done():
	}
}

const (
	exitFailure   = 1
	exitCancelled = 130
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

	b, err := readStorageFile(targetFS, u.Path)
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

	switch {
	case err == nil:
		return status, nil
	case errors.Is(err, context.Canceled):
		return exitCancelled, nil
	default:
		return exitFailure, err
	}
}

func readStorageFile(fs risoros.FS, path string) ([]byte, error) {
	const maxRetries = 3
	var b []byte
	var err error
	for i := 0; i < maxRetries+1; i++ {
		b, err = fs.ReadFile(path)
		if err == nil {
			break
		}
		time.Sleep(400 * time.Millisecond)
	}
	return b, err
}
