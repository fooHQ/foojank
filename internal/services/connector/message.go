package connector

import "github.com/nats-io/nats.go/micro"

type Message struct {
	req micro.Request
}

func (m Message) Data() []byte {
	return m.req.Data()
}

func (m Message) Reply(data []byte) error {
	return m.req.Respond(data)
}

func (m Message) ReplyError(code string, description string, data []byte) error {
	return m.req.Error(code, description, data)
}
