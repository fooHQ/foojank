package consumer

import (
	"github.com/nats-io/nats.go/jetstream"
)

type Message struct {
	msg jetstream.Msg
}

func NewMessage(msg jetstream.Msg) Message {
	return Message{
		msg: msg,
	}
}

func (m Message) Data() []byte {
	return m.msg.Data()
}

func (m Message) Ack() error {
	return m.msg.Ack()
}
