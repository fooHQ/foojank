package worker_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"testing"

	localfs "github.com/foohq/ren/filesystems/local"
	"github.com/nats-io/nats.go/jetstream"
	risoros "github.com/risor-io/risor/os"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/worker"
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

	eventCh := make(chan any, 2)

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
			Args:          []string{"arg1", "arg2"},
			Env:           []string{"TEST1", "hello", "TEST2", "world", "TEST3"},
			Connection:    js,
			Filesystems: map[string]risoros.FS{
				"file": fs,
			},
			EventCh: eventCh,
		}).Start(workerCtx)
		require.NoError(t, err)
	}()

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

	t.Run("check args", func(t *testing.T) {
		msg, err := msgs.Next()
		require.NoError(t, err)

		err = msg.Ack()
		require.NoError(t, err)

		v, err := proto.Unmarshal(msg.Data())
		require.NoError(t, err)
		require.IsType(t, proto.UpdateStdioLine{}, v)

		fields := strings.Fields(v.(proto.UpdateStdioLine).Text)
		require.Equal(t, []string{"args", "arg1", "arg2"}, fields)
	})

	t.Run("check env", func(t *testing.T) {
		msg, err := msgs.Next()
		require.NoError(t, err)

		err = msg.Ack()
		require.NoError(t, err)

		v, err := proto.Unmarshal(msg.Data())
		require.NoError(t, err)
		require.IsType(t, proto.UpdateStdioLine{}, v)

		fields := strings.Fields(v.(proto.UpdateStdioLine).Text)
		require.Contains(t, fields, "TEST1=hello")
		require.Contains(t, fields, "TEST2=world")
		require.Contains(t, fields, "TEST3=")
	})

	t.Run("check messages", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			in := fmt.Sprintf("input %d", i+1)
			b, err := proto.Marshal(proto.UpdateStdioLine{
				Text: in,
			})
			require.NoError(t, err)

			_, err = js.Publish(context.Background(), stdinName, b)
			require.NoError(t, err)

			// The same message is received twice - once for stdin, once for stdout.
			for y := 0; y < 2; y++ {
				msg, err := msgs.Next()
				require.NoError(t, err)

				err = msg.Ack()
				require.NoError(t, err)

				v, err := proto.Unmarshal(msg.Data())
				require.NoError(t, err)
				require.IsType(t, proto.UpdateStdioLine{}, v)

				out := v.(proto.UpdateStdioLine).Text
				require.Equal(t, in, out)
			}
		}
	})

	cancel()
	wg.Wait()

	require.Len(t, eventCh, 2)
	require.Equal(t, worker.EventWorkerStarted{ID: workerID}, <-eventCh)
	require.Equal(t, worker.EventWorkerStopped{ID: workerID}, <-eventCh)

	t.Run("check job status update", func(t *testing.T) {
		msg, err := msgs.Next()
		require.NoError(t, err)

		err = msg.Ack()
		require.NoError(t, err)

		v, err := proto.Unmarshal(msg.Data())
		require.NoError(t, err)
		require.IsType(t, proto.UpdateJob{}, v)
	})
}
