package reader

import "context"

type Message struct {
	ctx        context.Context
	data       string
	responseCh chan<- MessageResponse
}

func (m Message) Reply() error {
	select {
	case m.responseCh <- MessageResponse{}:
	case <-m.ctx.Done():
		return nil
	}
	return nil
}

func (m Message) Data() string {
	return m.data
}

type MessageResponse struct{}
