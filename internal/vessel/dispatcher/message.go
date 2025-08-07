package dispatcher

import (
	"context"

	"github.com/foohq/foojank/internal/vessel/connector"
)

type Message struct {
	msg  connector.Message
	data any
}

func (m Message) Ack() error {
	return m.msg.Ack()
}

func (m Message) Reply(ctx context.Context, data any) error {
	return m.msg.Reply(ctx, data)
}

func (m Message) Data() any {
	return m.data
}
