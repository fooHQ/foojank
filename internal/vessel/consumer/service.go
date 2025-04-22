package consumer

import (
	"context"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/vessel/log"
)

type Arguments struct {
	Servers           []string
	ConnectionOptions []nats.Option
	OutputCh          chan<- Message
}

type Service struct {
	args Arguments
}

func (s *Service) Start(ctx context.Context) error {
	servers := strings.Join(s.args.Servers, ",")

	// TODO: handle context closed!
	for {
		nc, err := nats.Connect(servers, s.args.ConnectionOptions...)
		if err != nil {
			log.Debug("cannot connect to the server", "error", err)
			continue
		}

		for !nc.IsConnected() {
			select {
			case <-time.After(3 * time.Second):
			case <-ctx.Done():
				return nil
			}
		}

		jetStream, err := jetstream.New(nc)
		if err != nil {
			// TODO: close connection and wait for reconnect
			return err
		}

		// TODO: use actual stream and consumer names from the args!
		c, err := jetStream.Consumer(ctx, "STREAM-TODO", "CONSUMER-TODO")
		if err != nil {
			// TODO: close connection and wait for reconnect
			return err
		}

		batch, err := c.FetchNoWait(5250)
		if err != nil {
			// TODO: close connection and wait for reconnect
			return err
		}

		for msg := range batch.Messages() {
			if msg == nil {
				break
			}
		}

	}
}
