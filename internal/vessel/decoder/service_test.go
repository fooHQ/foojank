package decoder_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/connector"
	"github.com/foohq/foojank/internal/vessel/decoder"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/foohq/foojank/proto"
)

func TestService(t *testing.T) {
	inputCh := make(chan connector.Message)
	outputCh := make(chan decoder.Message)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := decoder.New(decoder.Arguments{
			InputCh:  inputCh,
			OutputCh: outputCh,
		}).Start(ctx)
		require.NoError(t, err)
	}()

	responseCh := make(chan []byte)

	{
		b := []byte("_data_")
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		respMsg := <-responseCh
		require.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidMessage)))
	}

	{
		b, err := proto.NewCreateWorkerRequest()
		require.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-outputCh
		require.IsType(t, decoder.CreateWorkerRequest{}, outMsg.Data())
		err = outMsg.Reply(decoder.CreateWorkerResponse{
			ID: 1,
		})
		require.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		require.NoError(t, err)
		require.IsType(t, proto.CreateWorkerResponse{}, parsed)
		require.EqualValues(t, 1, parsed.(proto.CreateWorkerResponse).ID)
	}

	{
		b, err := proto.NewDestroyWorkerRequest(1)
		require.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-outputCh
		require.IsType(t, decoder.DestroyWorkerRequest{}, outMsg.Data())
		require.EqualValues(t, 1, outMsg.Data().(decoder.DestroyWorkerRequest).ID)
		err = outMsg.Reply(decoder.DestroyWorkerResponse{})
		require.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		require.NoError(t, err)
		require.IsType(t, proto.DestroyWorkerResponse{}, parsed)
	}

	{
		b, err := proto.NewGetWorkerRequest(1)
		require.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-outputCh
		require.IsType(t, decoder.GetWorkerRequest{}, outMsg.Data())
		require.EqualValues(t, 1, outMsg.Data().(decoder.GetWorkerRequest).ID)
		err = outMsg.Reply(decoder.GetWorkerResponse{
			ServiceName: "test",
			ServiceID:   "test-id",
		})
		require.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		require.NoError(t, err)
		require.IsType(t, proto.GetWorkerResponse{}, parsed)
		require.EqualValues(t, "test", parsed.(proto.GetWorkerResponse).ServiceName)
		require.EqualValues(t, "test-id", parsed.(proto.GetWorkerResponse).ServiceID)
	}

	{
		b, err := proto.NewDummyRequest()
		require.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		respMsg := <-responseCh
		require.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidAction)))
	}
}
