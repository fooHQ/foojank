package connector

import (
	"context"
)

type Message struct {
	msg  consumerMessage
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
