package decoder

import (
	"bytes"
	"context"
	"github.com/foojank/foojank/internal/services/vessel/errcodes"
	"github.com/foojank/foojank/internal/services/vessel/worker/connector"
	"github.com/foojank/foojank/internal/testutils"
	"github.com/foojank/foojank/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService(t *testing.T) {
	inputCh := make(chan connector.Message)
	dataCh := make(chan Message)
	stdinCh := make(chan Message)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := New(Arguments{
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
		assert.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidMessage)))
	}

	{
		b, err := proto.NewExecuteRequest([]byte("print"))
		assert.NoError(t, err)
		req := testutils.Request{
			FSubject:   "data",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		outMsg := <-dataCh
		assert.IsType(t, ExecuteRequest{}, outMsg.Data())
		err = outMsg.Reply(ExecuteResponse{
			Code: 1,
		})
		assert.NoError(t, err)

		b = <-responseCh
		parsed, err := proto.ParseResponse(b)
		assert.NoError(t, err)
		assert.IsType(t, proto.ExecuteResponse{}, parsed)
		assert.EqualValues(t, 1, parsed.(proto.ExecuteResponse).Code)
	}

	{
		b, err := proto.NewDummyRequest()
		assert.NoError(t, err)
		req := testutils.Request{
			FSubject:   "data",
			FData:      b,
			ResponseCh: responseCh,
		}
		msg := connector.NewMessage(req)
		inputCh <- msg
		respMsg := <-responseCh
		assert.True(t, bytes.HasPrefix(respMsg, []byte(errcodes.ErrInvalidAction)))
	}
}
