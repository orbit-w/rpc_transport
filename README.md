# MM RPC
MM RPC(Multi-Senders:Multi-Receivers) is a Go RPC(Remote procedure call) library for Golang.

Suitable for remote calls between services within the cluster,
it can also be used for message communication between intranet services

The StreamTransport transport layer relies on TCP underlying connections to provide reliability and ordering
and to provide stream-oriented multiplexing.

## Benchmark
```
➜  test git:(master) go test -v -run=^$ -bench .
goos: darwin
goarch: arm64
pkg: github.com/orbit-w/mmrpc/test
Benchmark_Call
Benchmark_Call-8                    	   42051	     27242 ns/op
Benchmark_Call_Concurrency
Benchmark_Call_Concurrency-8        	  244790	      4985 ns/op
Benchmark_Shoot
Benchmark_Shoot-8                   	12651248	        94.77 ns/op
Benchmark_Shoot_Concurrency
Benchmark_Shoot_Concurrency-8       	 5256571	       240.8 ns/op
Benchmark_AsyncCall
Benchmark_AsyncCall-8               	 1066633	       988.8 ns/op
Benchmark_AsyncCall_Concurrency
Benchmark_AsyncCall_Concurrency-8   	 1000000	      1211 ns/op
PASS
ok  	github.com/orbit-w/mmrpc/test	9.350s

```

## Client
```go
package main

import (
	"context"
	"github.com/orbit-w/mmrpc/rpc"
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

	_, err = cli.Call(context.Background(), 100, []byte{1})
	if err != nil {
		panic(err.Error())
	}
	
	if err = cli.AsyncCallC(100, []byte{3}, 100, func(ctx any, in []byte, err error) error {
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
	"github.com/orbit-w/mmrpc/rpc"
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