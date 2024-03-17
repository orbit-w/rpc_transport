package benchmark

import (
	"context"
	"github.com/orbit-w/rpc_transport/rpc_transport"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

const (
	host = "127.0.0.1:6800"
)

var (
	once      = new(sync.Once)
	rpcServer rpc_transport.RpcServer
)

func Benchmark_Call(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc_transport.Dial("node_00", "node_01", host)
	assert.NoError(b, err)
	ctx := context.Background()
	msg := []byte{2}
	b.ResetTimer()
	b.StartTimer()
	defer b.StopTimer()
	for i := 0; i < b.N; i++ {
		if _, err = cli.Call(ctx, msg); err != nil {
			b.Fatal(err.Error())
		}
	}
}

func Benchmark_Call_Concurrency(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc_transport.Dial("node_00", "node_01", host)
	assert.NoError(b, err)
	ctx := context.Background()
	b.StartTimer()
	defer b.StopTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err = cli.Call(ctx, []byte{2}); err != nil {
				b.Fatal(err.Error())
			}
		}
	})
}

func Benchmark_Shoot(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc_transport.Dial("node_00", "node_01", host)
	assert.NoError(b, err)
	msg := []byte{2}
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err = cli.Shoot(msg); err != nil {
			b.Fatal(err.Error())
		}
	}
	b.StopTimer()
}

func Benchmark_Shoot_Concurrency(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc_transport.Dial("node_00", "node_01", host)
	assert.NoError(b, err)
	msg := []byte{2}
	b.ResetTimer()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err = cli.Shoot(msg); err != nil {
				b.Fatal(err.Error())
			}
		}
	})

	b.StopTimer()
}

func Benchmark_AsyncCall(b *testing.B) {
	StartServe(b, nil)
	rpc_transport.SetInvokeCB(func(ctx any, in []byte, err error) error {
		return nil
	})
	cli, err := rpc_transport.Dial("node_00", "node_01", host)
	assert.NoError(b, err)
	pid := int64(100)
	msg := []byte{2}
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err = cli.AsyncCall(msg, pid); err != nil {
			b.Fatal(err.Error())
		}
	}
}

func Benchmark_AsyncCall_Concurrency(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc_transport.Dial("node_00", "node_01", host)
	assert.NoError(b, err)
	rpc_transport.SetInvokeCB(func(ctx any, in []byte, err error) error {
		return nil
	})
	msg := []byte{3}
	pid := int64(100)
	b.StartTimer()
	defer b.StopTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = cli.AsyncCall(msg, pid)
		}
	})
}

func StartServe(b *testing.B, rh rpc_transport.RequestHandle) {
	once.Do(func() {
		err := rpcServer.Serve(host, rh)
		assert.NoError(b, err)
	})
}
