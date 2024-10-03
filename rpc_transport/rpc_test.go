package rpc_transport

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"runtime/debug"
	"sync"
	"testing"
	"time"
)

const (
	local = "127.0.0.1:6800"
)

var (
	ServeOnce sync.Once

	rpcServer RpcServer
)

func Test_RPCShoot(t *testing.T) {
	Serve(t)

	cli, err := Dial("node_00", "node_01", local)
	assert.NoError(t, err)
	for i := 0; i < 1000; i++ {
		err = cli.Shoot([]byte{1})
		assert.NoError(t, err)
	}
	time.Sleep(time.Second * 2)
}

func Test_RPCCall(t *testing.T) {
	Serve(t)

	cli, err := Dial("node_00", "node_01", local)
	assert.NoError(t, err)

	for i := 0; i < 1000; i++ {
		in, err := cli.Call(context.Background(), []byte{1})
		assert.NoError(t, err)
		log.Println(in[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), RpcTimeout)
	_, err = cli.Call(ctx, []byte{1})
	assert.NoError(t, err)
	cancel()
	time.Sleep(time.Second * 5)
}

func Test_RPCCallWithCancel(t *testing.T) {
	Serve(t)

	cli, err := Dial("node_00", "node_01", local)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), RpcTimeout)
	go func() {
		_, err = cli.Call(ctx, []byte{1})
		fmt.Println("complete")
	}()
	assert.NoError(t, err)
	cancel()
	fmt.Println("cancel")
	time.Sleep(time.Second * 2)
}

func TestAsyncCall(t *testing.T) {
	Serve(t)

	cli, err := Dial("node_00", "node_01", local)
	assert.NoError(t, err)

	pid := int64(100)
	err = cli.AsyncCallC([]byte{1}, pid, func(ctx any, in []byte, err error) error {
		v := ctx.(int64)
		log.Println(v)
		log.Println("err: ", err)
		log.Println(in)
		return nil
	})
	assert.NoError(t, err)
	time.Sleep(time.Second * 15)
}

func TestBenchAsyncCall(t *testing.T) {
	Serve(t)

	pid := int64(100)
	msg := []byte{3}
	cli, err := Dial("node_00", "node_01", local)
	assert.NoError(t, err)
	SetInvokeCB(func(ctx any, in []byte, err error) error {
		if len(in) > 0 {
			fmt.Println(in[0])
		}
		return nil
	})
	for i := 0; i < 100000; i++ {
		if err := cli.AsyncCall(msg, pid); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(time.Second * 15)
}

func TestBenchAsyncCallPanic(t *testing.T) {
	Serve(t)

	pid := int64(100)
	msg := []byte{3}
	cli, err := Dial("node_00", "node_01", local)
	assert.NoError(t, err)
	SetInvokeCB(func(ctx any, in []byte, err error) error {
		panic("test")
		return nil
	})
	if err := cli.AsyncCall(msg, pid); err != nil {
		t.Error(err.Error())
	}
	time.Sleep(time.Second * 5)
}

func TestAsyncCallTimeout(t *testing.T) {
	Serve(t)
	SetReqHandle(func(req IRequest) error {
		r := req.NewReader()
		data, _ := io.ReadAll(r)
		switch data[0] {
		case 3:
			return nil
		default:
			return req.Response([]byte{1})
		}
	})

	cli, err := Dial("node_00", "node_01", "127.0.0.1:6800")
	assert.NoError(t, err)

	pid := int64(100)
	err = cli.AsyncCallC([]byte{3}, pid, func(ctx any, in []byte, err error) error {
		v := ctx.(int64)
		log.Println(v)
		log.Println("err: ", err)
		log.Println(in)
		return nil
	})
	assert.NoError(t, err)
	time.Sleep(time.Second * 15)
}

func Test_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r.(string))
			log.Println("stack: ", string(debug.Stack()))
		}
	}()

	panic("invalid pointer")
}

func Serve(t *testing.T) {
	ServeOnce.Do(func() {
		err := rpcServer.Serve(local, nil)
		assert.NoError(t, err)
	})
}

func Test_Misc(t *testing.T) {
	fmt.Println(NewRpcError(ErrDisconnect))
	fmt.Println(NewRpcError(nil))
	SetInvokeCB(func(ctx any, in []byte, err error) error {
		return nil
	})
}

func Test_RPCServerStop(t *testing.T) {
	var (
		s    RpcServer
		host = "127.0.0.1:6990"
	)
	err := s.Serve(host, nil)
	assert.NoError(t, err)

	cli, err := Dial("node_02", "node_01", host)
	err = cli.AsyncCall([]byte{1}, context.Background())
	assert.NoError(t, err)

	time.Sleep(time.Second)
	cli.Close()

	err = cli.AsyncCall([]byte{}, context.Background())
	assert.Error(t, err)
	fmt.Println(err)

	err = cli.AsyncCallC([]byte{}, context.Background(), nil)
	assert.Error(t, err)
	fmt.Println(err)

	_, err = cli.Call(context.Background(), nil)
	assert.Error(t, err)
	fmt.Println(err)

	err = cli.Shoot(nil)
	assert.Error(t, err)
	fmt.Println(err)

	assert.NoError(t, s.Stop())
}

func Test_AsyncInvoker(t *testing.T) {
	in := NewAsyncInvoker(nil, nil)
	fmt.Println(in.Invoke([]byte{}, errors.New("test")))
}
