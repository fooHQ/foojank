package agent

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type Message struct {
	msg      jetstream.Msg
	ID       string
	Seq      uint64
	Subject  string
	AgentID  string
	Sent     time.Time
	Received time.Time
}

func (m *Message) Data() []byte {
	return m.msg.Data()
}
