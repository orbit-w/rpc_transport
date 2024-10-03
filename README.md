# Rpc_Transport
RPC_Transport (Multi-Senders:Multi-Receivers) is a Go RPC(Remote procedure call) transport library for Golang.

RPC_Transport provides an efficient remote network communication module available in RPC mode

Suitable for remote calls between services within the cluster,
it can also be used for message communication between intranet services

The StreamTransport transport layer relies on TCP underlying connections to provide reliability and ordering
and to provide stream-oriented multiplexing.

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/orbit-w/rpc_transport/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/orbit-w/rpc_transport/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/orbit-w/mmrpc)](https://goreportcard.com/report/github.com/orbit-w/mmrpc)
[![codecov](https://codecov.io/gh/orbit-w/rpc_transport/graph/badge.svg?token=N86ZXD9S1E)](https://codecov.io/gh/orbit-w/rpc_transport)

# 中文：
RPC_Transport 可以为 RPC(Remote procedure call) 提供高效可靠底层远程通信传输链路，可以帮助开发者快速搭建RPC服务。
底层通信协议支持TCP,依赖TCP底层连接来提供可靠性和排序.
RPC_Transport 在面对高并发的情况下，性能也很优秀，底层避免多数无用的锁碰撞

支持 Shoot, Call, AsyncCall 三种消息通信模式

***Shoot***：单向远程投递消息，不关心回复

***Call***：单向远程投递消息，同阻塞式等待接收者的回复

***AsyncCall***: 单向远程投递消息，异步等待接收者的回复；当收到回复，系统会自动通知发送者调用回调函数
#

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/orbit-w/rpc_transport/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/orbit-w/rpc_transport/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/orbit-w/mmrpc)](https://goreportcard.com/report/github.com/orbit-w/mmrpc)
[![codecov](https://codecov.io/gh/orbit-w/rpc_transport/graph/badge.svg?token=N86ZXD9S1E)](https://codecov.io/gh/orbit-w/rpc_transport)

## Benchmark
```
go test ./benchmark/... -v -run=^-benchmem -bench=.
goos: darwin
goarch: arm64
pkg: github.com/orbit-w/rpc_transport/benchmark
Benchmark_Call
Benchmark_Call-8                    	   42486	     26653 ns/op
Benchmark_Call_Concurrency
Benchmark_Call_Concurrency-8        	  232701	      5103 ns/op
Benchmark_Shoot
Benchmark_Shoot-8                   	21751640	        55.48 ns/op
Benchmark_Shoot_Concurrency
Benchmark_Shoot_Concurrency-8       	10518703	       142.2 ns/op
Benchmark_AsyncCall
Benchmark_AsyncCall-8               	 2416693	       498.5 ns/op
Benchmark_AsyncCall_Concurrency
Benchmark_AsyncCall_Concurrency-8   	 2789655	       626.1 ns/op
PASS
ok  	github.com/orbit-w/rpc_transport/benchmark	11.343s
```

## Client
```go
package main

import (
	"context"
	"github.com/orbit-w/rpc_transport"
	"log"
)

/*
   @Author: orbit-w
   @File: main
   @2023 12月 周三 23:07
*/

func InitClient() {
	host := "127.0.0.1:6900"
	cli, err := rpc.NewClient("node_00", "node_01", host)
	if err != nil {
		panic(err.Error())
	}

	_, err = cli.Call(context.Background(), []byte{1})
	if err != nil {
		panic(err.Error())
	}
	
	if err = cli.AsyncCallC([]byte{3}, 100, func(ctx any, in []byte, err error) error {
		v := ctx.(int64)
		log.Println(v)
		log.Println("err: ", err)
		log.Println(in)
	}); err != nil {
		panic(err.Error())
	}
}

```

## Server
```go
package main

import (
	"github.com/orbit-w/rpc_transport"
)

/*
   @Author: orbit-w
   @File: main
   @2023 12月 周三 23:07
*/
func RunServer() {
	host := "127.0.0.1:6900"
	err := rpc.Serve(host, func(req rpc.IRequest) error {
		_ = req.Response([]byte{1})
		req.Return()
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
}
```