package processor

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/internal/engine"
	"github.com/foohq/foojank/internal/engine/os"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/decoder"
)

type Arguments struct {
	DataCh     <-chan decoder.Message
	StdinCh    <-chan decoder.Message
	StdoutCh   chan<- []byte
	Repository *repository.Client
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	group, _ := errgroup.WithContext(ctx)
	stdout := os.NewPipe()
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

	stdin := os.NewPipe()
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
			data := msg.Data()
			switch v := data.(type) {
			case decoder.ExecuteRequest:
				log.Debug("before load package", "repository", v.Repository, "path", v.FilePath)
				file, err := s.args.Repository.GetFile(ctx, v.Repository, v.FilePath)
				if err != nil {
					log.Debug(err.Error())
					_ = msg.ReplyError(errcodes.ErrRepositoryGetFile, err.Error(), "")
					continue
				}
				log.Debug("after load package", "repository", v.Repository, "path", v.FilePath)

				osCtx := os.NewContext(
					ctx,
					os.WithStdin(stdin),
					os.WithStdout(stdout),
					os.WithExitHandler(func(code int) {
						log.Debug("on exit", "code", code)
						cancel()
					}),
				)
				code, err := engine.CompilePackage(osCtx, file, int64(file.Size))
				if err != nil {
					log.Debug(err.Error())
					_ = msg.ReplyError(errcodes.ErrEngineCompile, err.Error(), "")
					continue
				}

				log.Debug("before run")
				err = code.Run(osCtx)
				if err != nil {
					log.Debug(err.Error())
					_ = msg.ReplyError(errcodes.ErrEngineRun, err.Error(), "")
					continue
				}
				log.Debug("after run")

				_ = msg.Reply(decoder.ExecuteResponse{
					Code: 0,
				})
			}

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
