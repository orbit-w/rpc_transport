# MM RPC
MM RPC(Multi-Senders:Multi-Receivers) is a RPC(Remote procedure call) library for Golang.

Suitable for remote calls between services within the cluster,
it can also be used for message communication between intranet services

The StreamTransport transport layer relies on TCP underlying connections to provide reliability and ordering
and to provide stream-oriented multiplexing.

## Benchmark
```
[16:26:39] [master ✖] ❱❱❱ go test -v -run=^$ -bench .
goos: darwin
goarch: arm64
pkg: github.com/orbit-w/mmrpc/test
Benchmark_Call
Benchmark_Call/benchmark_call
Benchmark_Call/benchmark_call-8 	   39750	     28449 ns/op
Benchmark_Call_Concurrency
Benchmark_Call_Concurrency-8    	  222679	      5283 ns/op
Benchmark_Shoot
Benchmark_Shoot/benchmark_call
Benchmark_Shoot/benchmark_call-8         	 4491601	       248.5 ns/op
Benchmark_Shoot_Concurrency
Benchmark_Shoot_Concurrency-8            	 4980124	       240.8 ns/op
PASS
ok  	github.com/orbit-w/mmrpc/test	12.935s
```