package publisher

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/message"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type NATSPublisherConfig struct {
	Connection *nats.Conn
}

type NATSPublisher struct {
	conf   NATSPublisherConfig
	logger *log.Logger
}

func NewNATSPublisher(logger *log.Logger, conf NATSPublisherConfig) *NATSPublisher {
	return &NATSPublisher{
		conf:   conf,
		logger: logger,
	}
}

func (p *NATSPublisher) Publish(ctx context.Context, msg message.Msg) error {
	data, err := protodaemon.Marshal(msg.Data())
	if err != nil {
		// TODO. check why error is not reported!
		return err
	}

	err = p.conf.Connection.PublishMsg(&nats.Msg{
		Subject: msg.Subject(),
		Data:    data,
	})
	if err != nil {
		return err
	}

	return nil
}
