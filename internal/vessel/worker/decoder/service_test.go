package decoder_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/internal/vessel/worker/connector"
	"github.com/foohq/foojank/internal/vessel/worker/decoder"
	"github.com/foohq/foojank/proto"
)

func TestService(t *testing.T) {
	inputCh := make(chan connector.Message)
	dataCh := make(chan decoder.Message)
	stdinCh := make(chan decoder.Message)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := decoder.New(decoder.Arguments{
			InputCh:     inputCh,
			DataSubject: "data",
			DataCh:      dataCh,
			StdinCh:     stdinCh,
		}).Start(ctx)
		assert.NoError(t, err)
	}()

	responseCh := make(chan []byte)

	{
		b := []byte("_data_")
		req := testutils.Request{
			FSubject:   "data",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		respMsg := <-responseCh
		require.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidMessage)))
	}

	{
		b, err := proto.NewExecuteRequest("/scripts/test.fzz", []string{"/scripts/test.fzz"})
		require.NoError(t, err)
		req := testutils.Request{
			FSubject:   "data",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-dataCh
		require.IsType(t, decoder.ExecuteRequest{}, outMsg.Data())
		data := outMsg.Data().(decoder.ExecuteRequest)
		require.Equal(t, []string{"/scripts/test.fzz"}, data.Args)
		require.Equal(t, "/scripts/test.fzz", data.FilePath)
		err = outMsg.Reply(decoder.ExecuteResponse{
			Code: 1,
		})
		require.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		require.NoError(t, err)
		require.IsType(t, proto.ExecuteResponse{}, parsed)
		require.EqualValues(t, 1, parsed.(proto.ExecuteResponse).Code)
	}

	{
		b, err := proto.NewDummyRequest()
		require.NoError(t, err)
		req := testutils.Request{
			FSubject:   "data",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		respMsg := <-responseCh
		require.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidAction)))
	}
}
