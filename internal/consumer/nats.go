package consumer

import (
	"context"
	"iter"

	"github.com/nats-io/nats.go"

	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/message"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type NATSConsumerConfig struct {
	Connection *nats.Conn
}

type NATSConsumer struct {
	conf   NATSConsumerConfig
	logger *log.Logger
}

func NewNATSConsumer(logger *log.Logger, conf NATSConsumerConfig) *NATSConsumer {
	return &NATSConsumer{
		conf:   conf,
		logger: logger,
	}
}

func (s *NATSConsumer) Messages(ctx context.Context) iter.Seq2[message.Msg, error] {
	var subs []*nats.Subscription
	defer func() {
		for _, sub := range subs {
			_ = sub.Unsubscribe()
		}
	}()

	subjects := []string{
		protodaemon.CreateUserSubject(),
		protodaemon.GetUserSubject(),
		protodaemon.ListUsersSubject(),
		protodaemon.CreateAgentSubject(),
		protodaemon.GetAgentSubject(),
		protodaemon.ListAgentsSubject(),
		protodaemon.IssueJWTSubject("*"),
	}
	msgCh := make(chan *nats.Msg, 2048)
	for _, subject := range subjects {
		sub, err := s.conf.Connection.ChanSubscribe(subject, msgCh)
		if err != nil {
			return func(yield func(message.Msg, error) bool) {
				yield(nil, err)
			}
		}
		subs = append(subs, sub)
	}

	return func(yield func(message.Msg, error) bool) {
		for {
			select {
			case msg := <-msgCh:
				s.logger.InfoContext(ctx, "Received a message: %s", msg.Subject)

				data, err := protodaemon.Unmarshal(msg.Data)
				if err != nil {
					continue
				}

				yield(NATSConsumerMessage{
					msg:  msg,
					data: data,
				}, nil)

			case <-ctx.Done():
				return
			}
		}
	}
}

type NATSConsumerMessage struct {
	msg  *nats.Msg
	data any
}

func (m NATSConsumerMessage) ID() string {
	return m.msg.Header.Get(nats.MsgIdHdr)
}

func (m NATSConsumerMessage) Subject() string {
	return m.msg.Subject
}

func (m NATSConsumerMessage) ReplySubject() string {
	return m.msg.Reply
}

func (m NATSConsumerMessage) Data() any {
	return m.data
}

func (m NATSConsumerMessage) Ack() error {
	return m.msg.Ack()
}
