package vessel_test

import (
	"testing"
)

// FIXME: test is broken and temporarily disabled

// TestService has the following steps:
// 1. publish messages to subject.
// 2. start the service
// 3. check that all messages were delivered
// 4. respond to messages
// 5. check that replies were published (requires another consumer)
func TestService(t *testing.T) {
	/*
		streamName := fmt.Sprintf("TEST-STREAM-%s", rand.Text())
		consumerName := fmt.Sprintf("TEST-CONSUMER-%s", rand.Text())
		subjectName := fmt.Sprintf("TEST.COMMANDS-%s", rand.Text())
		objectStoreName := fmt.Sprintf("TEST-OBJECT-STORE-%s", rand.Text())
		_, js := testutils.NewJetStreamConnection(t)

		stream, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
			Name:     streamName,
			Subjects: []string{subjectName},
		})
		require.NoError(t, err)

		consumer, err := js.CreateConsumer(context.Background(), streamName, jetstream.ConsumerConfig{
			Durable:       consumerName,
			DeliverPolicy: jetstream.DeliverLastPolicy,
			AckPolicy:     jetstream.AckExplicitPolicy,
		})
		require.NoError(t, err)

		store, err := js.CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
			Bucket: objectStoreName,
		})
		require.NoError(t, err)

		tests := []struct {
			request any
			reply   any
		}{
			{
				request: proto.StartWorkerRequest{
					File: filepath.Join(os.TempDir(), "test.zip"),
					Args: []string{"arg1", "arg2"},
					Env:  []string{"TEST", "hello"},
				},
				reply: proto.StartWorkerResponse{},
			},
		}

		for _, test := range tests {
			// 1. build file!
			//

			b, err := proto.Marshal(test.request)
			require.NoError(t, err)

			_, err = js.Publish(context.Background(), subjectName, b)
			require.NoError(t, err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := vessel.New(vessel.Arguments{
				ID:          "",
				Connection:  js,
				Stream:      stream,
				Consumer:    consumer,
				ObjectStore: store,
				Templates:   map[int]string{}, // TODO
			}).Start(ctx)
			require.NoError(t, err)
		}()

		for i, req := range requests {
			msg := <-outputCh
			require.NoError(t, msg.Ack())

			data := msg.Data()
			require.IsType(t, req, data)
			require.Equal(t, req, data)

			err := msg.Reply(context.Background(), replies[i])
			require.NoError(t, err)
		}

		c, err := js.OrderedConsumer(context.Background(), streamName, jetstream.OrderedConsumerConfig{})
		require.NoError(t, err)

		messages := append(requests, replies...)
		for i := 0; i < len(messages); i++ {
			msg, err := c.Next()
			require.NoError(t, err)

			err = msg.Ack()
			require.NoError(t, err)

			actual, err := proto.Unmarshal(msg.Data())
			require.NoError(t, err)

			require.Equal(t, nil, actual)
		}

		cancel()
		wg.Wait()*/
}
