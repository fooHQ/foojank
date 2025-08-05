package connector_test

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/connector"
	"github.com/foohq/foojank/proto"
)

// TestService has the following steps:
// 1. publish messages to subject.
// 2. start the service
// 3. check that all messages were delivered
// 4. respond to messages
// 5. check that replies were published (requires another consumer)
func TestService(t *testing.T) {
	streamName := fmt.Sprintf("TEST-STREAM-%s", rand.Text())
	consumerName := fmt.Sprintf("TEST-CONSUMER-%s", rand.Text())
	subjectName := fmt.Sprintf("TEST.COMMANDS-%s", rand.Text())
	srv, js := testutils.NewJetStreamConnection(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectName},
	})
	require.NoError(t, err)

	_, err = js.CreateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: jetstream.DeliverLastPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	require.NoError(t, err)

	requests := []any{
		proto.CreateJobRequest{
			Command: "exec",
			Args:    []string{"arg1", "arg2"},
			Env:     []string{"TEST", "hello"},
		},
		proto.CancelJobRequest{
			JobID: "job-123",
		},
	}
	replies := []any{
		proto.CreateJobResponse{
			JobID:         "job-123",
			StdinSubject:  "stdin-subject",
			StdoutSubject: "stdout-subject",
			Error:         errors.New("some error"),
		},
		proto.CancelJobResponse{
			Error: errors.New("some error"),
		},
	}

	for _, req := range requests {
		b, err := proto.Marshal(req)
		require.NoError(t, err)

		_, err = js.Publish(ctx, subjectName, b)
		require.NoError(t, err)
	}

	outputCh := make(chan connector.Message, len(requests))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := connector.New(connector.Arguments{
			Servers:  []string{srv.ClientURL()},
			Stream:   streamName,
			Consumer: consumerName,
			Subject:  subjectName,
			OutputCh: outputCh,
		}).Start(ctx)
		assert.NoError(t, err)
	}()

	for i, req := range requests {
		msg := <-outputCh
		require.NoError(t, msg.Ack())

		data := msg.Data()
		require.IsType(t, req, data)
		require.Equal(t, req, data)

		err := msg.Reply(ctx, replies[i])
		require.NoError(t, err)
	}

	c, err := js.OrderedConsumer(ctx, streamName, jetstream.OrderedConsumerConfig{})
	require.NoError(t, err)

	batch, err := c.FetchNoWait(len(requests) + len(replies))
	require.NoError(t, err)

	messages := append(requests, replies...)
	var i int
	for msg := range batch.Messages() {
		if msg == nil {
			break
		}
		actual, err := proto.Unmarshal(msg.Data())
		require.NoError(t, err)
		require.Equal(t, messages[i], actual)
		i++
	}

	cancel()
	wg.Wait()
}
