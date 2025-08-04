package proto_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto"
)

func TestParseCreateJobRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data, err := proto.NewCreateJobRequest("echo", []string{"-n", "hello"}, []string{"FOO=bar"})
		require.NoError(t, err)

		result, err := proto.ParseAction(data)
		require.NoError(t, err)

		req, ok := result.(proto.CreateJobRequest)
		require.True(t, ok)
		assert.Equal(t, "echo", req.Command)
		assert.Equal(t, []string{"-n", "hello"}, req.Args)
		assert.Equal(t, []string{"FOO=bar"}, req.Env)
	})

	t.Run("invalid message", func(t *testing.T) {
		_, err := proto.ParseAction([]byte("invalid"))
		require.Error(t, err)
	})
}

func TestParseCreateJobResponse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data, err := proto.NewCreateJobResponse("job-123", "stdin-subject", "stdout-subject", "some error")
		require.NoError(t, err)

		result, err := proto.ParseResponse(data)
		require.NoError(t, err)

		resp, ok := result.(proto.CreateJobResponse)
		require.True(t, ok)
		assert.Equal(t, "job-123", resp.JobID)
		assert.Equal(t, "stdin-subject", resp.StdinSubject)
		assert.Equal(t, "stdout-subject", resp.StdoutSubject)
		assert.Equal(t, errors.New("some error"), resp.Error)
	})

	t.Run("invalid message", func(t *testing.T) {
		_, err := proto.ParseResponse([]byte("invalid"))
		require.Error(t, err)
	})
}

func TestParseCancelJobRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data, err := proto.NewCancelJobRequest("job-123")
		require.NoError(t, err)

		result, err := proto.ParseAction(data)
		require.NoError(t, err)

		req, ok := result.(proto.CancelJobRequest)
		require.True(t, ok)
		assert.Equal(t, "job-123", req.JobID)
	})

	t.Run("invalid message", func(t *testing.T) {
		_, err := proto.ParseAction([]byte("invalid"))
		require.Error(t, err)
	})
}

func TestParseCancelJobResponse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data, err := proto.NewCancelJobResponse("some error")
		require.NoError(t, err)

		result, err := proto.ParseResponse(data)
		require.NoError(t, err)

		resp, ok := result.(proto.CancelJobResponse)
		require.True(t, ok)
		assert.Equal(t, errors.New("some error"), resp.Error)
	})

	t.Run("invalid message", func(t *testing.T) {
		_, err := proto.ParseResponse([]byte("invalid"))
		require.Error(t, err)
	})
}

func TestParseUpdateJob(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data, err := proto.NewUpdateJob("job-123", 42)
		require.NoError(t, err)

		result, err := proto.ParseResponse(data)
		require.NoError(t, err)

		update, ok := result.(proto.UpdateJob)
		require.True(t, ok)
		assert.Equal(t, "job-123", update.JobID)
		assert.Equal(t, int64(42), update.ExitStatus)
	})

	t.Run("invalid message", func(t *testing.T) {
		_, err := proto.ParseResponse([]byte("invalid"))
		require.Error(t, err)
	})
}

func TestParseDummyRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data, err := proto.NewDummyRequest()
		require.NoError(t, err)

		result, err := proto.ParseAction(data)
		require.NoError(t, err)

		_, ok := result.(proto.DummyRequest)
		require.True(t, ok)
	})

	t.Run("invalid message", func(t *testing.T) {
		_, err := proto.ParseAction([]byte("invalid"))
		require.Error(t, err)
	})
}
