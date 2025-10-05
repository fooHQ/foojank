package worker

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/foohq/ren"
	"github.com/foohq/ren/modules"
	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/message"
)

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

	err := forwardMessage(s.args.EventCh, EventWorkerStarted{
		WorkerID: s.args.ID,
	})
	if err != nil {
		log.Debug("Cannot forward a message", "error", err)
		return err
	}

	termCh := make(chan struct{})

	var wg sync.WaitGroup

	stdin := ren.NewPipe()
	stdinWriterCtx, stdinWriterCancel := context.WithCancel(ctx)
	stdinWriterCancelWrapper := func() {
		_ = stdin.Close()
		stdinWriterCancel()
	}
	defer stdinWriterCancelWrapper()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := stdinWriter(stdinWriterCtx, s.args.StdinCh, stdin)
		if err != nil {
			log.Debug("Stdin writer failed", "error", err)
		}
		termCh <- struct{}{}
	}()

	stdout := ren.NewPipe()
	stdoutReaderCtx, stdoutReaderCancel := context.WithCancel(ctx)
	stdoutReaderCancelWrapper := func() {
		_ = stdout.Close()
		stdoutReaderCancel()
	}
	defer stdoutReaderCancelWrapper()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := stdoutReader(stdoutReaderCtx, s.args.ID, stdout, s.args.EventCh)
		if err != nil {
			log.Debug("Stdout reader failed", "error", err)
		}
		termCh <- struct{}{}
	}()

	runnerCtx, runnerCancel := context.WithCancel(ctx)
	defer runnerCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		code, err := run(runnerCtx, s.args.File, s.args.Args, s.args.Env, stdin, stdout, s.args.Filesystems)
		if err != nil {
			log.Debug("Runner failed", "error", err)
		}

		err = forwardMessage(s.args.EventCh, EventWorkerStopped{
			WorkerID: s.args.ID,
			Status:   code,
			Error:    err,
		})
		if err != nil {
			log.Debug("Cannot forward a message", "error", err)
		}

		termCh <- struct{}{}
	}()

	cancels := []context.CancelFunc{
		runnerCancel,
		stdinWriterCancelWrapper,
		stdoutReaderCancelWrapper,
	}

	select {
	case <-ctx.Done():
		for _, cancel := range cancels {
			cancel()
			<-termCh
		}
	case <-termCh:
		// If an error occurs in one of the services, cancel all services without waiting for them to finish.
		// Some messages may be lost in the process.
		for _, cancel := range cancels {
			cancel()
		}
	}

	wg.Wait()

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
	const maxRetries = 5
	var b []byte
	var err error
	for i := 0; i < maxRetries+1; i++ {
		b, err = fs.ReadFile(path)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return b, err
}

func forwardMessage(outputCh chan<- message.Msg, msg message.Msg) error {
	select {
	case outputCh <- msg:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("timeout")
	}
}
