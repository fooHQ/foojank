package consumer

import (
	"context"
	"iter"

	"github.com/nats-io/nats.go"

	protogw "github.com/foohq/foojank/proto/gateway"

	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/message"
)

type Config struct {
	Connection *nats.Conn
	GatewayID  string
}

type Consumer struct {
	conf   Config
	logger *log.Logger
	msgCh  chan *nats.Msg
	subs   []*nats.Subscription
}

func New(logger *log.Logger, conf Config) *Consumer {
	return &Consumer{
		conf:   conf,
		logger: logger,
		msgCh:  make(chan *nats.Msg, 2048),
	}
}

func (s *Consumer) Subscribe(_ context.Context) error {
	subjects := []string{
		protogw.RegisterAgentSubject(s.conf.GatewayID),
		protogw.UnregisterAgentSubject(s.conf.GatewayID),
		protogw.CreateJobSubject(s.conf.GatewayID),
		protogw.CancelJobSubject(s.conf.GatewayID),
		protogw.JobOutputSubject(s.conf.GatewayID),
		protogw.JobInputSubject(s.conf.GatewayID),
	}

	for _, subject := range subjects {
		sub, err := s.conf.Connection.ChanSubscribe(subject, s.msgCh)
		if err != nil {
			return err
		}
		s.subs = append(s.subs, sub)
	}

	return nil
}

func (s *Consumer) Unsubscribe(_ context.Context) error {
	var errRes error
	for _, sub := range s.subs {
		err := sub.Unsubscribe()
		if err != nil {
			errRes = err
		}
	}
	return errRes
}

func (s *Consumer) Messages(ctx context.Context) iter.Seq2[message.Msg, error] {
	return func(yield func(message.Msg, error) bool) {
		for {
			select {
			case msg := <-s.msgCh:
				s.logger.InfoContext(ctx, "Received a message: %s", msg.Subject)

				data, err := protogw.Unmarshal(msg.Data)
				if err != nil {
					continue
				}

				yield(Message{
					msg:  msg,
					data: data,
				}, nil)

			case <-ctx.Done():
				return
			}
		}
	}
}

type Message struct {
	msg  *nats.Msg
	data any
}

func (m Message) ID() string {
	return m.msg.Header.Get(nats.MsgIdHdr)
}

func (m Message) Subject() string {
	return m.msg.Subject
}

func (m Message) ReplySubject() string {
	return m.msg.Reply
}

func (m Message) Data() any {
	return m.data
}

func (m Message) Ack() error {
	return m.msg.Ack()
}
