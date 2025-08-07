package worker

import (
	"context"
	"errors"
	"maps"
	"net/url"
	"sync"
	"time"

	"github.com/foohq/ren/builtins"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/ren"
	"github.com/foohq/ren/modules"
	renos "github.com/foohq/ren/os"

	"github.com/foohq/foojank/internal/repository"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/dispatcher"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/decoder"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	Filesystems map[string]risoros.FS
	Repository  *repository.Repository
	InputCh     <-chan dispatcher.Message
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
	// TODO: send start signal to the dispatcher.
	// TODO: send termination signal to the dispatcher.

	// Acquire the initial message which should always be CreateJobRequest and
	// use the Reply method to communicate with the connector.
	startMsg := <-s.args.InputCh

	stdout := renos.NewPipe()
	stdin := renos.NewPipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debug("started reading from stdout")
		defer log.Debug("stopped reading from stdout")
		b := make([]byte, 4096)
		for {
			n, err := stdout.Read(b)
			if err != nil {
				break
			}

			// TODO: use StdioLineUpdate message
			_ = startMsg.Reply(ctx, "TODO")
		}
		return
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debug("started reading from stdin")
		defer log.Debug("stopped reading from stdin")
		for {
			select {
			case msg := <-s.args.InputCh:
				log.Debug("before input", "value", string(b))
				_, _ = stdin.Write(b)
				log.Debug("after input")

			case <-ctx.Done():
				return nil
			}
		}
	}()

	data, ok := startMsg.Data().(proto.CreateJobRequest)
	if !ok {
		err := errors.New("initial command is not CreateJobRequest")
		log.Debug("cannot start worker", "error", err)
		return err
	}

	u, err := url.Parse(data.Command)
	if err != nil {
		log.Debug("cannot parse command URI", "error", err.Error())
		return err
	}

	log.Debug("before load package", "path", u.Path)

	b, err := s.readRepositoryFile(u.Path)
	if err != nil {
		log.Debug("cannot load package from the repository", "error", err.Error())
		return err
	}

	log.Debug("after load package", "path", u.Path)

	err = engineCompileAndRunPackage(
		ctx,
		b,
		renos.WithArgs(data.Args),
		renos.WithStdin(stdin),
		renos.WithStdout(stdout),
		renos.WithEnvVar("SERVICE_NAME", config.ServiceName), // TODO: add additional envs!
		renos.WithFilesystems(s.args.Filesystems),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Debug(err.Error())
		return err
	}

	log.Debug("closing stdin")
	_ = stdin.Close()
	log.Debug("closing stdout")
	_ = stdout.Close()

	log.Debug("waiting for all goroutines to stop")
	wg.Wait()
	log.Debug("all goroutines were stopped")
	return ctx.Err()
}

func (s *Service) readRepositoryFile(name string) ([]byte, error) {
	const retry = 3
	var b []byte
	var err error
	for i := 0; i < retry+1; i++ {
		b, err = s.args.Repository.ReadFile(name)
		// If there was no error break out from the loop and continue.
		// Otherwise, make another attempt to read the file.
		if err == nil {
			break
		}
		time.Sleep(400 * time.Millisecond)
	}
	return b, err
}

func engineCompileAndRunPackage(ctx context.Context, b []byte, opts ...renos.Option) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	exitHandler := func(code int) {
		log.Debug("on exit", "code", code)
		cancel()
	}
	opts = append(opts, renos.WithExitHandler(exitHandler))

	ros := renos.New(opts...)

	globals := make(map[string]any)
	maps.Copy(globals, modules.Globals())
	maps.Copy(globals, builtins.Globals())

	log.Debug("before run")

	err := ren.RunBytes(
		ctx,
		b,
		ros,
		ren.WithGlobals(globals),
	)
	if err != nil {
		return err
	}

	log.Debug("after run")

	return nil
}
