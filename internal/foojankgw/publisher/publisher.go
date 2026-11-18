package publisher

import (
	"context"

	"github.com/nats-io/nats.go"

	protogw "github.com/foohq/foojank/proto/gateway"

	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/message"
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
	data, err := protogw.Marshal(msg.Data())
	if err != nil {
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
