package processor

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"

	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/engine"
	engineos "github.com/foohq/foojank/internal/engine/os"
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
	stdout := engineos.NewPipe()
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

	stdin := engineos.NewPipe()
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

			b, err := s.args.Repository.ReadFile(v.FilePath)
			if err != nil {
				log.Debug(err.Error())
				_ = msg.ReplyError(errcodes.ErrRepositoryGetFile, err.Error(), "")
				continue
			}

			log.Debug("after load package", "path", v.FilePath)

			err = engineCompileAndRunPackage(
				ctx,
				b,
				engineos.WithArgs(v.Args),
				engineos.WithStdin(stdin),
				engineos.WithStdout(stdout),
				engineos.WithEnvVar("SERVICE_NAME", config.ServiceName),
				engineos.WithFilesystems(s.args.Filesystems),
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

func engineCompileAndRunPackage(ctx context.Context, b []byte, opts ...engineos.Option) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}

	exitHandler := func(code int) {
		log.Debug("on exit", "code", code)
		cancel()
	}
	opts = append(opts, engineos.WithExitHandler(exitHandler))

	log.Debug("before run")

	err = engine.Run(
		ctx,
		zr,
		engine.WithOS(engineos.New(opts...)),
		engine.WithGlobals(config.Modules()),
		engine.WithGlobals(config.Builtins()),
	)
	if err != nil {
		return err
	}

	log.Debug("after run")

	return nil
}
