package processor

import (
	"context"
	"errors"
	"maps"
	"time"

	"github.com/foohq/ren/builtins"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/ren"
	"github.com/foohq/ren/modules"
	renos "github.com/foohq/ren/os"

	"github.com/foohq/foojank/internal/repository"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/decoder"
)

type Arguments struct {
	DataCh      <-chan decoder.Message
	StdinCh     <-chan decoder.Message
	StdoutCh    chan<- []byte
	Repository  *repository.Repository
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
	group, _ := errgroup.WithContext(ctx)
	stdout := renos.NewPipe()
	group.Go(func() error {
		log.Debug("started reading from stdout")
		defer log.Debug("stopped reading from stdout")
		for {
			b := make([]byte, 4096)
			_, err := stdout.Read(b)
			if err != nil {
				break
			}

			select {
			case s.args.StdoutCh <- b:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	stdin := renos.NewPipe()
	group.Go(func() error {
		log.Debug("started reading from stdin")
		defer log.Debug("stopped reading from stdin")
		for {
			select {
			case msg := <-s.args.StdinCh:
				b := msg.Data().([]byte)
				log.Debug("before input", "value", string(b))
				_, _ = stdin.Write(b)
				log.Debug("after input")

			case <-ctx.Done():
				return nil
			}
		}
	})

loop:
	for {
		select {
		case msg := <-s.args.DataCh:
			v, ok := msg.Data().(decoder.ExecuteRequest)
			if !ok {
				continue
			}

			log.Debug("before load package", "path", v.FilePath)

			b, err := s.readRepositoryFile(v.FilePath)
			if err != nil {
				log.Debug(err.Error())
				_ = msg.ReplyError(errcodes.ErrRepositoryGetFile, err.Error(), "")
				continue
			}

			log.Debug("after load package", "path", v.FilePath)

			err = engineCompileAndRunPackage(
				ctx,
				b,
				renos.WithArgs(v.Args),
				renos.WithStdin(stdin),
				renos.WithStdout(stdout),
				renos.WithEnvVar("SERVICE_NAME", config.ServiceName),
				renos.WithFilesystems(s.args.Filesystems),
			)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Debug(err.Error())
				_ = msg.ReplyError(errcodes.ErrEngineRun, err.Error(), "")
				continue
			}

			_ = msg.Reply(decoder.ExecuteResponse{
				Code: 0,
			})

		case <-ctx.Done():
			break loop
		}
	}

	log.Debug("cleaning up")
	log.Debug("closing stdin")
	_ = stdin.Close()
	log.Debug("closing stdout")
	_ = stdout.Close()

	log.Debug("waiting for all goroutines to finish")
	_ = group.Wait()
	log.Debug("all goroutines finished")
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
