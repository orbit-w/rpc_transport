package test

import (
	"context"
	"github.com/orbit-w/mmrpc/rpc"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_RPCCall(t *testing.T) {
	err := rpc.Serve("127.0.0.1:6800", nil)
	assert.NoError(t, err)

	cli, err := rpc.NewClient("node_00", "node_01", "127.0.0.1:6800")
	assert.NoError(t, err)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	_, err = cli.Call(ctx, 100, []byte{1})
	assert.NoError(t, err)
	cli.Close()

	time.Sleep(time.Second * 5)
}
