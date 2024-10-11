package processor

import (
	"context"
	"github.com/foojank/foojank/internal/engine/os"
	"github.com/foojank/foojank/internal/log"
	"github.com/foojank/foojank/internal/services/vessel/worker/decoder"
	"github.com/risor-io/risor"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	DataCh   <-chan decoder.Message
	StdinCh  <-chan decoder.Message
	StdoutCh chan<- []byte
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
				src := string(v.Data)
				log.Debug("before eval")
				_, err := risor.Eval(ctx, src)
				log.Debug("after eval")
				if err != nil {
					_, _ = stdout.Write([]byte(err.Error()))
				}

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
