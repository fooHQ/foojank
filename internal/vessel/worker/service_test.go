package worker_test

import (
	"testing"
)

// FIXME: test is broken and temporarily disabled

func TestService(t *testing.T) {
	/*
		jobID := rand.Text()
		streamName := fmt.Sprintf("TEST-STREAM-%s", rand.Text())
		stdinName := fmt.Sprintf("TEST-STDIN-%s", rand.Text())
		stdoutName := fmt.Sprintf("TEST-STDOUT-%s", rand.Text())
		workManagerName := fmt.Sprintf("TEST-COMMANDS-%s", rand.Text())
		_, js := testutils.NewJetStreamConnection(t)

		stream, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
			Name:     streamName,
			Subjects: []string{stdinName, stdoutName, workManagerName},
		})
		require.NoError(t, err)

		fs, err := localfs.NewFS()
		require.NoError(t, err)

		workerCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eventCh := make(chan any, 2)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := worker.New(worker.Arguments{
				JobID:              jobID,
				Stream:             stream,
				StdinSubject:       stdinName,
				StdoutSubject:      stdoutName,
				WorkManagerSubject: workManagerName,
				Entrypoint:         "./testdata/test.zip",
				Args:               []string{"arg1", "arg2"},
				Env:                []string{"TEST1", "hello", "TEST2", "world", "TEST3"},
				Connection:         js,
				Filesystems: map[string]risoros.FS{
					"file": fs,
				},
				EventCh: eventCh,
			}).Start(workerCtx)
			require.NoError(t, err)
		}()

		var messages = []string{
			fmt.Sprintf("input 1"),
			fmt.Sprintf("input 2"),
			fmt.Sprintf("input 3"),
			fmt.Sprintf("input 4"),
			fmt.Sprintf("input 5"),
		}

		// Produce messages to stdin.
		for _, message := range messages {
			b, err := proto.Marshal(proto.UpdateWorkerStdio{
				Data: message,
			})
			require.NoError(t, err)

			_, err = js.Publish(context.Background(), stdinName, b)
			require.NoError(t, err)
		}

		c, err := js.CreateConsumer(context.Background(), streamName, jetstream.ConsumerConfig{
			Name:           jobID + "-check",
			DeliverPolicy:  jetstream.DeliverAllPolicy,
			AckPolicy:      jetstream.AckExplicitPolicy,
			MaxAckPending:  1,
			FilterSubjects: []string{stdoutName, workManagerName},
		})
		require.NoError(t, err)

		msgs, err := c.Messages()
		require.NoError(t, err)
		defer msgs.Stop()

		var output string
		for i := 0; i < len(messages)+2; i++ {
			msg, err := msgs.Next()
			require.NoError(t, err)

			err = msg.Ack()
			require.NoError(t, err)

			v, err := proto.Unmarshal(msg.Data())
			require.NoError(t, err)
			require.IsType(t, proto.UpdateWorkerStdio{}, v)
			output += v.(proto.UpdateWorkerStdio).Data
		}

		lines := strings.Split(output, "\n")

		t.Run("check args", func(t *testing.T) {
			require.Equal(t, "args arg1 arg2", lines[0])
		})

		t.Run("check env variables", func(t *testing.T) {
			fields := strings.Fields(lines[1])
			require.Contains(t, fields, "TEST1=hello")
			require.Contains(t, fields, "TEST2=world")
			require.Contains(t, fields, "TEST3=")
		})

		cancel()
		wg.Wait()

		require.Len(t, eventCh, 2)
		require.Equal(t, worker.EventWorkerStarted{ID: jobID}, <-eventCh)
		require.Equal(t, worker.EventWorkerStopped{ID: jobID}, <-eventCh)

		// Check job status update
		{
			msg, err := msgs.Next()
			require.NoError(t, err)

			err = msg.Ack()
			require.NoError(t, err)

			v, err := proto.Unmarshal(msg.Data())
			require.NoError(t, err)
			require.IsType(t, proto.UpdateWorkerStatus{}, v)
		}*/
}
