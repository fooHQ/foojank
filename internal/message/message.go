package message

type Msg interface {
	ID() string
	Subject() string
	ReplySubject() string
	Data() any
	Ack() error
}
