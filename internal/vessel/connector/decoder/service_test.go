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

	{
		b, err := proto.NewCreateJobRequest(
			"execute",
			[]string{"arg1", "arg2"},
			[]string{"TEST", "hello"},
		)
		require.NoError(t, err)

		inputCh <- consumer.NewMessage(testutils.NewMsg("test-subject", b), nil)
		outMsg := <-outputCh
		require.IsType(t, proto.CreateJobRequest{}, outMsg.Data())
	}

	{
		b, err := proto.NewCancelJobRequest("job-32")
		require.NoError(t, err)

		inputCh <- consumer.NewMessage(testutils.NewMsg("test-subject", b), nil)
		outMsg := <-outputCh
		require.IsType(t, proto.CancelJobRequest{}, outMsg.Data())
	}
}
