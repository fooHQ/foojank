package connector_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/testutils"
	"github.com/foohq/foojank/internal/vessel/worker/connector"
)

func TestService(t *testing.T) {
	var stdinSubject = "stdin"
	var dataSubject = "data"
	var reqData = []byte("_data_")
	var respData = []byte("_resp_data_")

	infoCh := make(chan connector.InfoMessage)
	outputCh := make(chan connector.Message)
	_, nc := testutils.NewNatsServerAndConnection(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := connector.New(connector.Arguments{
			Name:         "test",
			Version:      "0.0.1",
			StdinSubject: stdinSubject,
			DataSubject:  dataSubject,
			Connection:   nc,
			InfoCh:       infoCh,
			OutputCh:     outputCh,
		}).Start(ctx)
		assert.NoError(t, err)
	}()

	info := <-infoCh
	require.NotEmpty(t, info.ServiceID())
	require.Equal(t, "test", info.ServiceName())

	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := <-outputCh
			assert.Equal(t, reqData, msg.Data())

			err := msg.Reply(respData)
			assert.NoError(t, err)
		}()

		resp, err := nc.Request(dataSubject, reqData, 2*time.Second)
		require.NoError(t, err)
		require.Equal(t, respData, resp.Data)
	}

	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := <-outputCh
			assert.Equal(t, reqData, msg.Data())

			err := msg.Reply(respData)
			assert.NoError(t, err)
		}()

		resp, err := nc.Request(stdinSubject, reqData, 2*time.Second)
		require.NoError(t, err)
		require.Equal(t, respData, resp.Data)
	}

	cancel()
	wg.Wait()
}
