package connector

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Message struct {
	msg nats.Msg
}

func NewMessage(req micro.Request) Message {
	return Message{
		msg: req,
	}
}

func (m Message) Data() []byte {
	return m.msg.Data()
}

func (m Message) Reply(data []byte) error {
	return m.msg.Respond(data)
}

func (m Message) ReplyError(code string, description string, data []byte) error {
	return m.msg.Error(code, description, data)
}
