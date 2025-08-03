package consumer_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/consumer"
)

func TestService(t *testing.T) {
	streamName := "TEST-STREAM"
	consumerName := fmt.Sprintf("TEST-CONSUMER-%s", rand.Text())
	srv, js := testutils.NewJetStreamConnection(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"TEST.COMMANDS"},
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
			Servers:           []string{srv.ClientURL()},
			Stream:            "TEST-STREAM",
			Consumer:          "TEST-CONSUMER",
			BatchSize:         5,
			ReconnectInterval: 5 * time.Second,
			OutputCh:          outputCh,
		}).Start(ctx)
		require.NoError(t, err)
	}()

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("COMMAND %d", i+1)
		_, err = js.Publish(ctx, "TEST.COMMANDS", []byte(msg))
		require.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		msg := <-outputCh
		expectedData := fmt.Sprintf("COMMAND %d", i+1)
		require.Equal(t, expectedData, string(msg.Data()))
	}

	cancel()
	wg.Wait()
}
