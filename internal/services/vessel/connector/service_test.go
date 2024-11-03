package connector

import (
	"context"
	"github.com/foohq/foojank/internal/testutils"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	var rpcSubject = "rpc"
	var reqData = []byte("_data_")
	var respData = []byte("_resp_data_")

	outputCh := make(chan Message)
	_, nc := testutils.NewNatsServerAndConnection(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := New(Arguments{
			Name:       "test",
			Version:    "0.0.1",
			Metadata:   nil,
			RpcSubject: rpcSubject,
			Connection: nc,
			OutputCh:   outputCh,
		}).Start(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := <-outputCh
			assert.Equal(t, reqData, msg.Data())

			err := msg.Reply(respData)
			assert.NoError(t, err)
		}()

		resp, err := nc.Request(rpcSubject, reqData, 2*time.Second)
		assert.NoError(t, err)
		assert.Equal(t, respData, resp.Data)
	}

	cancel()
	wg.Wait()
}
