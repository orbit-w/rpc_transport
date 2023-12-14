package test

import (
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
	StartServe(b)
	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	time.Sleep(time.Second * 5)
	ctx := context.Background()
	msg := []byte{2}
	b.Run("benchmark call", func(b *testing.B) {
		b.ResetTimer()
		b.StartTimer()
		defer b.StopTimer()
		for i := 0; i < b.N; i++ {
			if _, err = cli.Call(ctx, 100, msg); err != nil {
				b.Fatal(err.Error())
			}
		}
	})
}

func Benchmark_Call_Concurrency(b *testing.B) {
	StartServe(b)

	cli, err := rpc.NewClient("node_00", "node_01", host)
	assert.NoError(b, err)
	ctx := context.Background()
	b.StartTimer()
	defer b.StopTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err = cli.Call(ctx, 100, []byte{2}); err != nil {
				b.Fatal(err.Error())
			}
		}
	})
}

func StartServe(b *testing.B) {
	once.Do(func() {
		err := rpc.Serve(host, nil)
		assert.NoError(b, err)
	})
}
