package consumer_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/connector/consumer"
)

func TestService(t *testing.T) {
	streamName := fmt.Sprintf("TEST-STREAM-%s", rand.Text())
	consumerName := fmt.Sprintf("TEST-CONSUMER-%s", rand.Text())
	subjectName := fmt.Sprintf("TEST.COMMANDS-%s", rand.Text())
	_, js := testutils.NewJetStreamConnection(t)

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

	outputCh := make(chan consumer.Message)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = consumer.New(consumer.Arguments{
			Connection: js,
			Stream:     streamName,
			Consumer:   consumerName,
			OutputCh:   outputCh,
		}).Start(ctx)
		require.NoError(t, err)
	}()

	var messages []string
	for i := 0; i < 10; i++ {
		messages = append(messages, fmt.Sprintf("COMMAND %d", i+1))
	}

	for _, msg := range messages {
		_, err = js.Publish(ctx, subjectName, []byte(msg))
		require.NoError(t, err)
	}

	for _, expected := range messages {
		msg := <-outputCh
		require.Equal(t, expected, string(msg.Data()))
	}

	cancel()
	wg.Wait()
}
