package decoder

import (
	"bytes"
	"context"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/services/vessel/connector"
	"github.com/foojank/foojank/internal/testutils"
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
		assert.True(t, bytes.HasPrefix(respMsg, []byte("error")))
	}

	{
		b, err := vessel.NewCreateWorkerRequest()
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
	}

	{
		b, err := vessel.NewDestroyWorkerRequest(1)
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
	}

	{
		b, err := vessel.NewGetWorkerRequest(1)
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
	}

	{
		b, err := vessel.NewDummyRequest()
		assert.NoError(t, err)
		req := testutils.Request{
			FSubject:   "test",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		respMsg := <-responseCh
		assert.True(t, bytes.HasPrefix(respMsg, []byte("error")))
	}
}
