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
		data, err := proto.NewCreateJobResponse("job-123", "stdin-subject", "stdout-subject", errors.New("some error"))
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
		data, err := proto.NewCancelJobResponse(errors.New("some error"))
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

func TestMarshalUnmarshal(t *testing.T) {
	testError := errors.New("test error")

	tests := []struct {
		name        string
		input       any
		want        any
		wantMarshal bool
		wantErr     error
	}{
		{
			name: "CreateJobRequest",
			input: proto.CreateJobRequest{
				Command: "test-cmd",
				Args:    []string{"arg1", "arg2"},
				Env:     []string{"VAR1", "val1", "VAR2", "val2"},
			},
			want: proto.CreateJobRequest{
				Command: "test-cmd",
				Args:    []string{"arg1", "arg2"},
				Env:     []string{"VAR1", "val1", "VAR2", "val2"},
			},
			wantMarshal: true,
		},
		{
			name: "CreateJobRequest with special characters",
			input: proto.CreateJobRequest{
				Command: "test",
				Args:    []string{"arg with spaces", "arg\nwith\nnewlines", "arg\twith\ttabs"},
				Env:     []string{"VAR1", "value with spaces", "VAR2", "value\nwith\nnewlines"},
			},
			want: proto.CreateJobRequest{
				Command: "test",
				Args:    []string{"arg with spaces", "arg\nwith\nnewlines", "arg\twith\ttabs"},
				Env:     []string{"VAR1", "value with spaces", "VAR2", "value\nwith\nnewlines"},
			},
			wantMarshal: true,
		},
		{
			name: "CreateJobRequest with empty strings",
			input: proto.CreateJobRequest{
				Command: "test",
				Args:    []string{"", "non-empty", ""},
				Env:     []string{"VAR1", "", "VAR2", ""},
			},
			want: proto.CreateJobRequest{
				Command: "test",
				Args:    []string{"", "non-empty", ""},
				Env:     []string{"VAR1", "", "VAR2", ""},
			},
			wantMarshal: true,
		},
		{
			name: "CreateJobRequest Unicode characters",
			input: proto.CreateJobRequest{
				Command: "测试",
				Args:    []string{"参数1", "パラメータ2", "매개변수3"},
				Env:     []string{"变量1", "值1", "変数2", "値2"},
			},
			want: proto.CreateJobRequest{
				Command: "测试",
				Args:    []string{"参数1", "パラメータ2", "매개변수3"},
				Env:     []string{"变量1", "值1", "変数2", "値2"},
			},
			wantMarshal: true,
		},
		{
			name: "CreateJobRequest with nil slices",
			input: proto.CreateJobRequest{
				Command: "test",
				Args:    nil,
				Env:     nil,
			},
			want: proto.CreateJobRequest{
				Command: "test",
				Args:    nil,
				Env:     nil,
			},
			wantMarshal: true,
		},
		{
			name: "CreateJobResponse with nil error",
			input: proto.CreateJobResponse{
				JobID:         "job-123",
				StdinSubject:  "stdin",
				StdoutSubject: "stdout",
				Error:         nil,
			},
			want: proto.CreateJobResponse{
				JobID:         "job-123",
				StdinSubject:  "stdin",
				StdoutSubject: "stdout",
				Error:         nil,
			},
			wantMarshal: true,
		},
		{
			name: "CancelJobResponse with nil error",
			input: proto.CancelJobResponse{
				Error: nil,
			},
			want: proto.CancelJobResponse{
				Error: nil,
			},
			wantMarshal: true,
		},
		{
			name: "CreateJobResponse without error",
			input: proto.CreateJobResponse{
				JobID:         "job-123",
				StdinSubject:  "stdin.123",
				StdoutSubject: "stdout.123",
			},
			want: proto.CreateJobResponse{
				JobID:         "job-123",
				StdinSubject:  "stdin.123",
				StdoutSubject: "stdout.123",
			},
			wantMarshal: true,
		},
		{
			name: "CreateJobResponse with error",
			input: proto.CreateJobResponse{
				JobID:         "job-123",
				StdinSubject:  "stdin.123",
				StdoutSubject: "stdout.123",
				Error:         testError,
			},
			want: proto.CreateJobResponse{
				JobID:         "job-123",
				StdinSubject:  "stdin.123",
				StdoutSubject: "stdout.123",
				Error:         testError,
			},
			wantMarshal: true,
		},
		{
			name: "CancelJobRequest",
			input: proto.CancelJobRequest{
				JobID: "job-123",
			},
			want: proto.CancelJobRequest{
				JobID: "job-123",
			},
			wantMarshal: true,
		},
		{
			name:        "CancelJobResponse without error",
			input:       proto.CancelJobResponse{},
			want:        proto.CancelJobResponse{},
			wantMarshal: true,
		},
		{
			name: "CancelJobResponse with error",
			input: proto.CancelJobResponse{
				Error: testError,
			},
			want: proto.CancelJobResponse{
				Error: testError,
			},
			wantMarshal: true,
		},
		{
			name: "UpdateJob",
			input: proto.UpdateJob{
				JobID:      "job-123",
				ExitStatus: 42,
			},
			want: proto.UpdateJob{
				JobID:      "job-123",
				ExitStatus: 42,
			},
			wantMarshal: true,
		},
		{
			name: "UpdateStdioLine",
			input: proto.UpdateStdioLine{
				Text: "text",
			},
			want: proto.UpdateStdioLine{
				Text: "text",
			},
			wantMarshal: true,
		},
		{
			name:    "Unsupported type",
			input:   struct{}{},
			wantErr: proto.ErrUnknownMessage,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: proto.ErrUnknownMessage,
		},
		{
			name:    "Empty input",
			input:   []byte{},
			wantErr: proto.ErrUnknownMessage,
		},
		{
			name:    "Invalid data",
			input:   []byte("invalid"),
			wantErr: proto.ErrUnknownMessage,
		},
		{
			name:    "Invalid data",
			input:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
			wantErr: proto.ErrUnknownMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Marshal
			marshaled, err := proto.Marshal(tt.input)
			if !tt.wantMarshal {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, marshaled)

			// Test Unmarshal
			unmarshaled, err := proto.Unmarshal(marshaled)
			require.NoError(t, err)
			require.Equal(t, tt.want, unmarshaled)
		})
	}
}
