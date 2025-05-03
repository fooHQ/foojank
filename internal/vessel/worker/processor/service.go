package processor

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/risor-io/risor"
	"github.com/risor-io/risor/vm"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/engine"
	engineos "github.com/foohq/foojank/internal/engine/os"
	"github.com/foohq/foojank/internal/uri"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/worker/decoder"
)

type Arguments struct {
	DataCh      <-chan decoder.Message
	StdinCh     <-chan decoder.Message
	StdoutCh    chan<- []byte
	URIHandlers map[string]engineos.URIHandler
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
			file, err := downloadFile(s.args.URIHandlers, v.FilePath)
			if err != nil {
				log.Debug(err.Error())
				_ = msg.ReplyError(errcodes.ErrRepositoryGetFile, err.Error(), "")
				continue
			}
			log.Debug("after load package", "path", v.FilePath)

			err = engineCompileAndRunPackage(ctx,
				file,
				engineos.WithArgs(v.Args),
				engineos.WithStdin(stdin),
				engineos.WithStdout(stdout),
				engineos.WithEnvVar("SERVICE_NAME", config.ServiceName),
				engineos.WithURIHandlers(s.args.URIHandlers),
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

func downloadFile(uriHandlers map[string]engineos.URIHandler, pth string) (*File, error) {
	u, err := uri.ToURL(pth)
	if err != nil {
		return nil, err
	}

	handler, ok := uriHandlers[u.Scheme]
	if !ok {
		return nil, engineos.ErrHandlerNotFound
	}

	fs, pth, err := handler.GetFS(u)
	if err != nil {
		return nil, err
	}

	f, err := fs.Open(pth)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return &File{
		b:        bytes.NewReader(b),
		Name:     pth,
		Size:     uint64(info.Size()),
		Modified: info.ModTime(),
	}, nil
}

func engineCompileAndRunPackage(ctx context.Context, file *File, opts ...engineos.Option) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	exitHandler := func(code int) {
		log.Debug("on exit", "code", code)
		cancel()
	}

	osCtx := engineos.NewContext(
		ctx,
		append(opts, engineos.WithExitHandler(exitHandler))...,
	)

	conf := risor.NewConfig(
		risor.WithoutDefaultGlobals(),
		risor.WithGlobals(config.Modules()),
		risor.WithGlobals(config.Builtins()),
	)

	importer, err := engineos.NewFzzImporter(file, int64(file.Size), conf.CompilerOpts()...)
	if err != nil {
		return err
	}

	log.Debug("before run")

	vmOpts := conf.VMOpts()
	vmOpts = append(vmOpts, vm.WithImporter(importer))
	err = engine.Run(osCtx, vmOpts...)
	if err != nil {
		return err
	}

	log.Debug("after run")

	return nil
}
