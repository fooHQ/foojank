package publisher

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/message"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type Config struct {
	Connection *nats.Conn
}

type Publisher struct {
	conf   Config
	logger *log.Logger
}

func New(logger *log.Logger, conf Config) *Publisher {
	return &Publisher{
		conf:   conf,
		logger: logger,
	}
}

func (p *Publisher) Publish(ctx context.Context, msg message.Msg) error {
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
