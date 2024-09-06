package connector

import "github.com/nats-io/nats.go/micro"

type Message struct {
	req micro.Request
}

func (m Message) Subject() string {
	return m.req.Subject()
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

type InfoMessage struct {
	serviceName string
	serviceID   string
}

func (m InfoMessage) ServiceName() string {
	return m.serviceName
}

func (m InfoMessage) ServiceID() string {
	return m.serviceID
}
