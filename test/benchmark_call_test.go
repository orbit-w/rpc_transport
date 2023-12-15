package test

import (
	"container/heap"
	"context"
	"github.com/orbit-w/mmrpc/rpc"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

const (
	host = "127.0.0.1:6800"
)

var (
	once = new(sync.Once)
)

func Benchmark_Call(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	ctx := context.Background()
	pid := int64(100)
	msg := []byte{2}
	b.Run("benchmark call", func(b *testing.B) {
		b.ResetTimer()
		b.StartTimer()
		defer b.StopTimer()
		for i := 0; i < b.N; i++ {
			if _, err = cli.Call(ctx, pid, msg); err != nil {
				b.Fatal(err.Error())
			}
		}
	})
}

func Benchmark_Call_Concurrency(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	ctx := context.Background()
	pid := int64(100)
	b.StartTimer()
	defer b.StopTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err = cli.Call(ctx, pid, []byte{2}); err != nil {
				b.Fatal(err.Error())
			}
		}
	})
}

func Benchmark_Shoot(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	time.Sleep(time.Second * 2)
	msg := []byte{2}
	pid := int64(100)
	b.Run("benchmark call", func(b *testing.B) {
		b.ResetTimer()
		b.StartTimer()
		defer b.StopTimer()
		for i := 0; i < b.N; i++ {
			if err = cli.Shoot(pid, msg); err != nil {
				b.Fatal(err.Error())
			}
		}
	})

	b.StopTimer()
}

func Benchmark_Shoot_Concurrency(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	msg := []byte{2}
	pid := int64(100)
	b.ResetTimer()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err = cli.Shoot(pid, msg); err != nil {
				b.Fatal(err.Error())
			}
		}
	})

	b.StopTimer()
}

func Benchmark_AsyncCall(b *testing.B) {
	StartPProf()
	heap.Fix()
	StartServe(b, nil)
	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	pid := int64(100)
	msg := []byte{2}
	b.Run("benchmark call", func(b *testing.B) {
		b.ResetTimer()
		b.StartTimer()
		defer b.StopTimer()
		for i := 0; i < b.N; i++ {
			if err = cli.AsyncCall(pid, msg, pid); err != nil {
				b.Fatal(err.Error())
			}
		}
	})
}

func Benchmark_AsyncCall_Concurrency(b *testing.B) {
	StartServe(b, nil)
	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	rpc.SetInvokeCB(func(ctx any, in []byte, err error) error {
		return nil
	})
	msg := []byte{3}
	pid := int64(100)
	b.StartTimer()
	defer b.StopTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = cli.AsyncCall(pid, msg, pid)
		}
	})
}

func StartServe(b *testing.B, rh rpc.RequestHandle) {
	once.Do(func() {
		err := rpc.Serve(host, rh)
		assert.NoError(b, err)
	})
}
