package message

import (
	"errors"
)

var (
	ErrUnsupported = errors.New("no response expected")
)

type Msg interface {
	ID() string
	Subject() string
	Data() any
	Ack() error
}
