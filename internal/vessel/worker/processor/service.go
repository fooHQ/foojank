package processor

import (
	"bytes"
	"context"
	"io"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/risor-io/risor"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/filesystems/filefs"
	"github.com/foohq/foojank/filesystems/natsfs"
	"github.com/foohq/foojank/internal/engine"
	engineos "github.com/foohq/foojank/internal/engine/os"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/decoder"
)

type Arguments struct {
	DataCh    <-chan decoder.Message
	StdinCh   <-chan decoder.Message
	StdoutCh  chan<- []byte
	JetStream jetstream.JetStream
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
	fileFs, err := filefs.New()
	if err != nil {
		log.Debug("cannot initialize filefs", "error", err)
		return err
	}

	store, err := s.args.JetStream.ObjectStore(ctx, config.ServiceName)
	if err != nil {
		log.Debug("cannot open object store", "error", err)
		return err
	}

	natsFs, err := natsfs.New(store)
	if err != nil {
		log.Debug("cannot initialize natsfs", "error", err)
		return err
	}

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

			log.Debug("before load package", "repository", v.Repository, "path", v.FilePath)
			file, err := fetchFile(ctx, s.args.JetStream, v.Repository, v.FilePath)
			if err != nil {
				log.Debug(err.Error())
				_ = msg.ReplyError(errcodes.ErrRepositoryGetFile, err.Error(), "")
				continue
			}
			log.Debug("after load package", "repository", v.Repository, "path", v.FilePath)

			err = engineCompileAndRunPackage(ctx,
				file,
				engineos.WithArgs(v.Args),
				engineos.WithStdin(stdin),
				engineos.WithStdout(stdout),
				engineos.WithEnvVar("SERVICE_NAME", config.ServiceName),
				engineos.WithURLHandler("file", "", fileFs),
				engineos.WithURLHandler("natsfs", "", natsFs),
			)
			if err != nil {
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

func fetchFile(ctx context.Context, js jetstream.JetStream, repository, filename string) (*File, error) {
	s, err := js.ObjectStore(ctx, repository)
	if err != nil {
		return nil, err
	}

	res, err := s.Get(ctx, filename)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	b, err := io.ReadAll(res)
	if err != nil {
		return nil, err
	}

	info, err := res.Info()
	if err != nil {
		return nil, err
	}

	return &File{
		b:        bytes.NewReader(b),
		Name:     info.Name,
		Size:     info.Size,
		Modified: info.ModTime,
	}, nil
}

func engineCompileAndRunPackage(ctx context.Context, file *File, opts ...engineos.Option) error {
	osCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts = append(opts, engineos.WithExitHandler(func(code int) {
		log.Debug("on exit", "code", code)
		cancel()
	}))

	osCtx = engineos.NewContext(
		osCtx,
		opts...,
	)

	risorOpts := []risor.Option{
		risor.WithoutDefaultGlobals(),
		risor.WithGlobals(config.Modules()),
		risor.WithGlobals(config.Builtins()),
	}
	code, err := engine.CompilePackage(osCtx, file, int64(file.Size), risorOpts...)
	if err != nil {
		return err
	}

	log.Debug("before run")
	err = code.Run(osCtx)
	if err != nil {
		return err
	}
	log.Debug("after run")

	return nil
}
