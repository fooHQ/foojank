package daemon

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Message struct {
	msg  jetstream.Msg
	meta *jetstream.MsgMetadata
}

func (m *Message) ID() string {
	return m.msg.Headers().Get(nats.MsgIdHdr)
}

func (m *Message) Subject() string {
	return m.msg.Subject()
}

func (m *Message) Timestamp() time.Time {
	return m.meta.Timestamp
}

func (m *Message) Data() []byte {
	return m.msg.Data()
}
