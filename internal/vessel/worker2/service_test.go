package worker_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"

	localfs "github.com/foohq/ren/filesystems/local"
	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/worker2"
	"github.com/foohq/foojank/proto"
)

func TestService(t *testing.T) {
	workerID := rand.Text()
	streamName := fmt.Sprintf("TEST-STREAM-%s", rand.Text())
	stdinName := fmt.Sprintf("TEST-STDIN-%s", rand.Text())
	stdoutName := fmt.Sprintf("TEST-STDOUT-%s", rand.Text())
	updateName := fmt.Sprintf("TEST-UPDATE-%s", rand.Text())
	_, js := testutils.NewJetStreamConnection(t)

	_, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{stdinName, stdoutName, updateName},
	})
	require.NoError(t, err)

	fs, err := localfs.NewFS()
	require.NoError(t, err)

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := worker.New(worker.Arguments{
			ID:            workerID,
			Stream:        streamName,
			StdinSubject:  stdinName,
			StdoutSubject: stdoutName,
			UpdateSubject: updateName,
			Entrypoint:    "./testdata/test.zip",
			// TODO: script should print args and env vars
			Args:       []string{"arg1", "arg2"},
			Env:        []string{"TEST1", "hello", "TEST2", "world", "TEST3"},
			Connection: js,
			Filesystems: map[string]risoros.FS{
				"file": fs,
			},
			EventCh: nil,
		}).Start(workerCtx)
		require.NoError(t, err)
	}()

	var messages [][]byte
	for i := 0; i < 10; i++ {
		b, err := proto.Marshal(proto.UpdateStdioLine{
			Text: fmt.Sprintf("input %d", i+1),
		})
		require.NoError(t, err)
		messages = append(messages, b)
	}

	for _, msg := range messages {
		_, err = js.Publish(context.Background(), stdinName, msg)
		require.NoError(t, err)
	}

	// Cancel context after sending all messages to trigger worker shutdown
	cancel()

	c, err := js.CreateConsumer(context.Background(), streamName, jetstream.ConsumerConfig{
		Name:          workerID + "-check",
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
	})
	require.NoError(t, err)

	msgs, err := c.Messages()
	require.NoError(t, err)
	defer msgs.Stop()

	var output string
	// TODO: fix loop count
	for i := 0; i < len(messages)*3; i++ {
		msg, err := msgs.Next()
		require.NoError(t, err)

		err = msg.Ack()
		require.NoError(t, err)

		v, err := proto.Unmarshal(msg.Data())
		require.NoError(t, err)

		output += v.(proto.UpdateStdioLine).Text
		println(output)
	}

	println("FINAL OUTPUT:")
	println("=============")
	println(output)
	println("=============")
	// TODO: check output

	wg.Wait()
}
