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

	{
		inputCh <- proto.CreateJobResponse{
			JobID:         "job-123",
			StdinSubject:  "stdin-subject",
			StdoutSubject: "stdout-subject",
			Error:         errors.New("some error"),
		}
		expected, err := proto.NewCreateJobResponse("job-123", "stdin-subject", "stdout-subject", errors.New("some error"))
		require.NoError(t, err)
		actual := <-outputCh
		require.Equal(t, expected, actual)
	}

	{
		inputCh <- proto.CancelJobResponse{
			Error: errors.New("some error"),
		}
		expected, err := proto.NewCancelJobResponse(errors.New("some error"))
		require.NoError(t, err)
		actual := <-outputCh
		require.Equal(t, expected, actual)
	}

	{
		inputCh <- proto.UpdateJob{
			JobID:      "job-123",
			ExitStatus: 1,
		}
		expected, err := proto.NewUpdateJob("job-123", 1)
		require.NoError(t, err)
		actual := <-outputCh
		require.Equal(t, expected, actual)
	}

	{
		inputCh <- "invalid message"
		require.Empty(t, outputCh)
	}
}
