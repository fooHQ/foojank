package writer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/proto"
)

type Arguments struct {
	Connection jetstream.JetStream
	File       risoros.File
	Subject    string
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
	log.Debug("Service started", "service", "vessel.workmanager.worker.writer")
	defer log.Debug("Service stopped", "service", "vessel.workmanager.worker.writer")

	fileReaderOutCh := make(chan []byte)
	encoderOutCh := make(chan []byte)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return fileReader(groupCtx, s.args.File, fileReaderOutCh)
	})

	group.Go(func() error {
		return encoder(groupCtx, fileReaderOutCh, encoderOutCh)
	})

	group.Go(func() error {
		return publisher(groupCtx, s.args.Connection, encoderOutCh, s.args.Subject)
	})

	<-groupCtx.Done()
	_ = s.args.File.Close()

	return group.Wait()
}

func fileReader(ctx context.Context, stdout risoros.File, outputCh chan<- []byte) error {
	log.Debug("Service started", "service", "vessel.workmanager.worker.writer.filereader")
	defer log.Debug("Service stopped", "service", "vessel.workmanager.worker.writer.filereader")

	b := make([]byte, 4096)
	for {
		n, err := stdout.Read(b)
		if err != nil {
			return nil
		}

		select {
		case outputCh <- b[:n]:
		case <-ctx.Done():
			return nil
		}
	}
}

func encoder(ctx context.Context, inputCh <-chan []byte, outputCh chan<- []byte) error {
	log.Debug("Service started", "service", "vessel.workmanager.worker.writer.encoder")
	defer log.Debug("Service stopped", "service", "vessel.workmanager.worker.writer.encoder")

	for {
		select {
		case msg := <-inputCh:
			data := proto.UpdateStdioLine{
				Text: string(msg),
			}

			encoded, err := proto.Marshal(data)
			if err != nil {
				log.Debug("Cannot encode message", "error", err)
				continue
			}

			select {
			case outputCh <- encoded:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func publisher(ctx context.Context, conn jetstream.JetStream, inputCh <-chan []byte, subject string) error {
	log.Debug("Service started", "service", "vessel.workmanager.worker.writer.publisher")
	defer log.Debug("Service stopped", "service", "vessel.workmanager.worker.writer.publisher")

	for {
		select {
		case msg := <-inputCh:
			_, err := conn.Publish(ctx, subject, msg)
			if err != nil {
				log.Debug("Cannot publish message", "error", err)
				continue
			}

		case <-ctx.Done():
			return nil
		}
	}
}
