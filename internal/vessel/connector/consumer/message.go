package consumer

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
)

type Message struct {
	msg     jetstream.Msg
	replyCh chan<- any
}

func NewMessage(msg jetstream.Msg, replyCh chan<- any) Message {
	return Message{
		msg:     msg,
		replyCh: replyCh,
	}
}

func (m Message) Data() []byte {
	return m.msg.Data()
}

func (m Message) Ack() error {
	return m.msg.Ack()
}

func (m Message) Reply(ctx context.Context, data any) error {
	if m.replyCh == nil {
		return errors.New("message does not expect a reply")
	}
	select {
	case m.replyCh <- data:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
