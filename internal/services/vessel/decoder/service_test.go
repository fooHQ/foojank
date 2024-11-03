package decoder

import (
	"bytes"
	"context"
	"github.com/foohq/foojank/internal/services/vessel/connector"
	"github.com/foohq/foojank/internal/services/vessel/errcodes"
	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService(t *testing.T) {
	inputCh := make(chan connector.Message)
	outputCh := make(chan Message)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := New(Arguments{
			InputCh:  inputCh,
			OutputCh: outputCh,
		}).Start(ctx)
		assert.NoError(t, err)
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
		assert.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidMessage)))
	}

	{
		b, err := proto.NewCreateWorkerRequest()
		assert.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-outputCh
		assert.IsType(t, CreateWorkerRequest{}, outMsg.Data())
		err = outMsg.Reply(CreateWorkerResponse{
			ID: 1,
		})
		assert.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		assert.NoError(t, err)
		assert.IsType(t, proto.CreateWorkerResponse{}, parsed)
		assert.EqualValues(t, 1, parsed.(proto.CreateWorkerResponse).ID)
	}

	{
		b, err := proto.NewDestroyWorkerRequest(1)
		assert.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-outputCh
		assert.IsType(t, DestroyWorkerRequest{}, outMsg.Data())
		assert.EqualValues(t, 1, outMsg.Data().(DestroyWorkerRequest).ID)
		err = outMsg.Reply(DestroyWorkerResponse{})
		assert.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		assert.NoError(t, err)
		assert.IsType(t, proto.DestroyWorkerResponse{}, parsed)
	}

	{
		b, err := proto.NewGetWorkerRequest(1)
		assert.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-outputCh
		assert.IsType(t, GetWorkerRequest{}, outMsg.Data())
		assert.EqualValues(t, 1, outMsg.Data().(GetWorkerRequest).ID)
		err = outMsg.Reply(GetWorkerResponse{
			ServiceName: "test",
			ServiceID:   "test-id",
		})
		assert.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		assert.NoError(t, err)
		assert.IsType(t, proto.GetWorkerResponse{}, parsed)
		assert.EqualValues(t, "test", parsed.(proto.GetWorkerResponse).ServiceName)
		assert.EqualValues(t, "test-id", parsed.(proto.GetWorkerResponse).ServiceID)
	}

	{
		b, err := proto.NewDummyRequest()
		assert.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		respMsg := <-responseCh
		assert.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidAction)))
	}
}
