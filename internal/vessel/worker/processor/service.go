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
	stdin := os.NewPipe()
	stdout := os.NewPipe()
	ros := os.New(ctx, stdin, stdout)
	ctx = ros.Context()

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error {
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
	group.Go(func() error {
		for {
			select {
			case msg := <-s.args.StdinCh:
				b := msg.Data().([]byte)
				log.Debug("before input: %q", string(b))
				_, _ = stdin.Write(b)
				log.Debug("after input")

			case <-ctx.Done():
				return nil
			}
		}
	})
	// Wait for context closure and then close stdin and stdout files.
	// This unblocks (cancels) any potential pending writes/reads to/from the files
	// and allows the service to shutdown gracefully.
	group.Go(func() error {
		<-ctx.Done()
		_ = stdin.Close()
		_ = stdout.Close()
		return nil
	})

loop:
	for {
		select {
		case msg := <-s.args.DataCh:
			data := msg.Data()
			switch v := data.(type) {
			case decoder.ExecuteRequest:
				log.Debug("before load package %s:%s", v.Repository, v.FilePath)
				file, err := s.args.Repository.GetFile(ctx, v.Repository, v.FilePath)
				if err != nil {
					log.Debug(err.Error())
					_ = msg.ReplyError(errcodes.ErrRepositoryGetFile, err.Error(), "")
					continue
				}
				log.Debug("after load package %s:%s", v.Repository, v.FilePath)

				eng, err := engine.Unpack(ctx, file, int64(file.Size))
				if err != nil {
					log.Debug(err.Error())
					_ = msg.ReplyError(errcodes.ErrEngineUnpack, err.Error(), "")
					continue
				}

				log.Debug("before run")
				_, err = eng.Run(ctx)
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

	return group.Wait()
}
