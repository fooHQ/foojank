package decoder

import (
	"context"
	"github.com/foojank/foojank/internal/services/vessel/connector"
)

type Message struct {
	ctx        context.Context
	req        connector.Message
	responseCh chan<- MessageResponse
	data       any
}

func (m Message) Reply(data any) error {
	select {
	case m.responseCh <- MessageResponse{
		req:  m.req,
		data: data,
	}:
	case <-m.ctx.Done():
		return nil
	}
	return nil
}

func (m Message) ReplyError(code string, description string, data any) error {
	select {
	case m.responseCh <- MessageResponse{
		req:  m.req,
		data: data,
		err: &MessageError{
			Code:        code,
			Description: description,
		},
	}:
	case <-m.ctx.Done():
		return nil
	}
	return nil
}

func (m Message) Data() any {
	return m.data
}

type MessageResponse struct {
	req  connector.Message
	data any
	err  *MessageError
}

type MessageError struct {
	Code        string
	Description string
}

func (m MessageResponse) Request() connector.Message {
	return m.req
}

func (m MessageResponse) Data() any {
	return m.data
}

func (m MessageResponse) Error() *MessageError {
	return m.err
}

type CreateWorkerRequest struct{}

type CreateWorkerResponse struct {
	ID uint64
}

type DestroyWorkerRequest struct {
	ID uint64
}

type DestroyWorkerResponse struct{}

type GetWorkerRequest struct {
	ID uint64
}

type GetWorkerResponse struct {
	ServiceID string
}
