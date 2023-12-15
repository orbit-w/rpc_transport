# MM RPC
MM RPC(Multi-Senders:Multi-Receivers) is a RPC(Remote procedure call) library for Golang.

Suitable for remote calls between services within the cluster,
it can also be used for message communication between intranet services

The StreamTransport transport layer relies on TCP underlying connections to provide reliability and ordering
and to provide stream-oriented multiplexing.

## Benchmark
```
➜  test git:(master) ✗ go test -v -run=^$ -bench .
goos: darwin
goarch: arm64
pkg: github.com/orbit-w/mmrpc/test
Benchmark_Call
Benchmark_Call/benchmark_call
Benchmark_Call/benchmark_call-8     	   40617	     28387 ns/op
Benchmark_Call_Concurrency
Benchmark_Call_Concurrency-8        	  226645	      5150 ns/op
Benchmark_Shoot
Benchmark_Shoot/benchmark_call
Benchmark_Shoot/benchmark_call-8    	 4058671	       263.8 ns/op
Benchmark_Shoot_Concurrency
Benchmark_Shoot_Concurrency-8       	 3804121	       306.0 ns/op
Benchmark_AsyncCall
Benchmark_AsyncCall/benchmark_call
Benchmark_AsyncCall/benchmark_call-8         	  925357	      1111 ns/op
Benchmark_AsyncCall_Concurrency
Benchmark_AsyncCall_Concurrency-8            	 1000000	      1497 ns/op
PASS
ok  	github.com/orbit-w/mmrpc/test	10.646s
```