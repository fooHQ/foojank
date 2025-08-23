package message

import "context"

type Msg interface {
	Ack() error
	Reply(context.Context, any) error
	Data() any
}
