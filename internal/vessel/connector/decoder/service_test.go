package decoder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/connector/consumer"
	"github.com/foohq/foojank/internal/vessel/connector/decoder"
	"github.com/foohq/foojank/proto"
)

func TestService(t *testing.T) {
	inputCh := make(chan consumer.Message)
	outputCh := make(chan decoder.Message)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := decoder.New(decoder.Arguments{
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
			name: "CreateJobRequest",
			input: proto.CreateJobRequest{
				Command: "test-cmd",
				Args:    []string{"arg1", "arg2"},
				Env:     []string{"VAR1", "val1", "VAR2", "val2"},
			},
		},
		{
			name: "CancelJobRequest",
			input: proto.CancelJobRequest{
				JobID: "job-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := proto.Marshal(tt.input)
			require.NoError(t, err)

			inMsg := testutils.NewMsg("test-subject", marshaled)
			inputCh <- consumer.NewMessage(inMsg, nil)
			outMsg := <-outputCh
			require.IsType(t, tt.input, outMsg.Data())
			require.Equal(t, tt.input, outMsg.Data())
		})
	}
}
