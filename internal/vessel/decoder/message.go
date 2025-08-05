package decoder

import (
	"github.com/foohq/foojank/internal/vessel/consumer"
)

type Message struct {
	msg  consumer.Message
	data any
}

func (m Message) Ack() error {
	return m.msg.Ack()
}

func (m Message) Data() any {
	return m.data
}
