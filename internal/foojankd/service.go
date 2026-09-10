package foojankd

import (
	"context"
	"errors"
	"iter"
	"sync"
	"time"

	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/message"
)

type Config struct {
	Consumer  Consumer
	Handler   Handler
	Publisher Publisher
}

type Service struct {
	args   Config
	logger *log.Logger
}

func New(logger *log.Logger, args Config) *Service {
	return &Service{
		args:   args,
		logger: logger,
	}
}

func (s *Service) Start(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Service %q started", "foojankd")
	defer s.logger.InfoContext(ctx, "Service %q stopped", "foojankd")

	// Capacity must be equal or greater than the total number of goroutines tracked by the WaitGroup.
	termCh := make(chan struct{}, 100)

	var wg sync.WaitGroup
	var cancels []context.CancelFunc

	{
		consumerOutCh := make(chan message.Msg)
		publisherInCh := make(chan message.Msg, 128)

		consumerCtx, consumerCancel := context.WithCancel(context.Background())
		cancels = append(cancels, consumerCancel)

		wg.Go(func() {
			err := consumer(consumerCtx, s.logger, s.args.Consumer, consumerOutCh)
			if err != nil {
				s.logger.ErrorContext(ctx, "NATS consumer error: %v", err)
			}
			termCh <- struct{}{}
		})

		handlerCtx, handlerCancel := context.WithCancel(context.Background())
		cancels = append(cancels, handlerCancel)

		wg.Go(func() {
			err := handler(handlerCtx, s.logger, s.args.Handler, consumerOutCh, publisherInCh)
			if err != nil {
				s.logger.ErrorContext(ctx, "Handler error: %v", err)
			}
			termCh <- struct{}{}
		})

		publisherCtx, publisherCancel := context.WithCancel(context.Background())
		cancels = append(cancels, publisherCancel)

		wg.Go(func() {
			err := publisher(publisherCtx, s.logger, s.args.Publisher, publisherInCh)
			if err != nil {
				s.logger.ErrorContext(ctx, "NATS publisher error: %v", err)
			}
			termCh <- struct{}{}
		})
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

type Consumer interface {
	Messages(context.Context) iter.Seq2[message.Msg, error]
}

func consumer(ctx context.Context, logger *log.Logger, consumer Consumer, outputCh chan message.Msg) error {
	logger.InfoContext(ctx, "Service %q started", "foojankd.consumer")
	defer logger.InfoContext(ctx, "Service %q stopped", "foojankd.consumer")

	for msg, err := range consumer.Messages(ctx) {
		if err != nil {
			logger.ErrorContext(ctx, "Cannot read a message: %v", err)
			return err
		}

		err = forwardMessage(outputCh, msg)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot forward a message: %v", err)
			continue
		}
	}

	return nil
}

type Publisher interface {
	Publish(context.Context, message.Msg) error
}

func publisher(ctx context.Context, logger *log.Logger, publisher Publisher, inputCh <-chan message.Msg) error {
	logger.InfoContext(ctx, "Service %q started", "foojankd.publisher")
	defer logger.InfoContext(ctx, "Service %q stopped", "foojankd.publisher")

	var exit bool
	var cancel context.CancelFunc

loop:
	for {
		select {
		case msg := <-inputCh:
			err := publisher.Publish(ctx, msg)
			if err != nil {
				logger.WarnContext(ctx, "Cannot publish a message to %q: %v", msg.Subject(), err)
				continue
			}

			_ = msg.Ack()

		case <-ctx.Done():
			if !exit {
				ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
				exit = true
				continue loop
			}
			break loop
		}
	}

	cancel()

	if len(inputCh) != 0 {
		logger.WarnContext(ctx, "Some messages were lost (%d messages)", len(inputCh))
	}

	return nil
}

type Handler interface {
	Match(message.Msg) (func(context.Context) message.Msg, bool)
}

func handler(ctx context.Context, logger *log.Logger, handler Handler, inputCh <-chan message.Msg, outputCh chan<- message.Msg) error {
	logger.InfoContext(ctx, "Service %q started", "foojankd.handler")
	defer logger.InfoContext(ctx, "Service %q stopped", "foojankd.handler")

	for {
		for {
			select {
			case msg := <-inputCh:
				fn, ok := handler.Match(msg)
				if !ok {
					_ = msg.Ack()
					continue
				}

				resp := fn(ctx)

				err := forwardMessage(outputCh, resp)
				if err != nil {
					logger.WarnContext(ctx, "Cannot forward a message: %v", err)
					continue
				}

			case <-ctx.Done():
				return nil
			}
		}
	}
}

func forwardMessage(outputCh chan<- message.Msg, msg message.Msg) error {
	select {
	case outputCh <- msg:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("timeout")
	}
}
