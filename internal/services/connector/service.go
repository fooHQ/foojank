package connector

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"strconv"
	"strings"
)

const (
	// TODO: pass as args!
	Name    string = "vessel"
	Version string = "0.1.0"
)

type Arguments struct {
	Connection *nats.Conn
	OutputCh   chan<- Message
}

type Message struct {
	Data       []byte
	Error      error
	ResponseCh chan<- Message
}

func (m *Message) Reply(ctx context.Context, data []byte) {
	msg := Message{
		Data: data,
	}

	select {
	case m.ResponseCh <- msg:
	case <-ctx.Done():
		return
	}
}

type Error struct {
	Code        int
	Description string
}

func (e *Error) Error() string {
	return strconv.Itoa(e.Code) + " " + e.Description
}

type Service struct {
	args Arguments
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	ms, err := micro.AddService(s.args.Connection, micro.Config{
		Name:    Name,
		Version: Version,
	})
	if err != nil {
		return err
	}
	defer ms.Stop()

	rpcSubject := strings.Join([]string{micro.APIPrefix, "RPC", Name, ms.Info().ID}, ".")
	err = ms.AddEndpoint("rpc", micro.ContextHandler(ctx, s.handler), micro.WithEndpointSubject(rpcSubject))
	if err != nil {
		return err
	}

	// TODO: add wg group wait!

	<-ctx.Done()
	return nil
}

func (s *Service) handler(ctx context.Context, req micro.Request) {
	// TODO: use wg to track goroutines!
	go func() {
		responseCh := make(chan Message, 1)
		msg := Message{
			Data:       req.Data(),
			ResponseCh: responseCh,
		}

		select {
		case s.args.OutputCh <- msg:
		case <-ctx.Done():
			return
		}

		select {
		case respMsg := <-responseCh:
			s.responseHandler(req, respMsg)
		case <-ctx.Done():
			return
		}
	}()
}

func (s *Service) responseHandler(req micro.Request, msg Message) {
	if err := msg.Error; err != nil {
		var errorMsg *Error
		if errors.As(err, &errorMsg) {
			code := strconv.Itoa(errorMsg.Code)
			_ = req.Error(code, errorMsg.Description, msg.Data)
		} else {
			_ = req.Error("500", err.Error(), msg.Data)
		}
		return
	}

	_ = req.Respond(msg.Data)
}
