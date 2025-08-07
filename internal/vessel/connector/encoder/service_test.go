package encoder_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/vessel/connector/encoder"
	"github.com/foohq/foojank/proto"
)

func TestService(t *testing.T) {
	inputCh := make(chan any)
	outputCh := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := encoder.New(encoder.Arguments{
			InputCh:  inputCh,
			OutputCh: outputCh,
		}).Start(ctx)
		assert.NoError(t, err)
	}()

	tests := []struct {
		name  string
		input any
	}{
		{
			name: "CreateJobResponse with error",
			input: proto.CreateJobResponse{
				JobID:         "job-123",
				StdinSubject:  "stdin.123",
				StdoutSubject: "stdout.123",
				Error:         errors.New("test error"),
			},
		},
		{
			name: "CancelJobResponse with error",
			input: proto.CancelJobResponse{
				Error: errors.New("test error"),
			},
		},
		{
			name: "UpdateJob",
			input: proto.UpdateJob{
				JobID:      "job-123",
				ExitStatus: 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCh <- tt.input
			outMsg := <-outputCh

			marshaled, err := proto.Marshal(tt.input)
			require.NoError(t, err)
			require.IsType(t, marshaled, outMsg)
			require.Equal(t, marshaled, outMsg)
		})
	}
}
