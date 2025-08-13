package publisher

type Message struct {
	subject string
	data    []byte
}

func NewMessage(subject string, data []byte) Message {
	return Message{
		subject: subject,
		data:    data,
	}
}
