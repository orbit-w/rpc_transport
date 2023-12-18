package transport

import (
	"errors"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

func Test_Transport(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host)

	conn := DialWithOps(host, &DialOption{
		RemoteNodeId:  "node_0",
		CurrentNodeId: "node_1",
	})
	defer func() {
		_ = conn.Close()
	}()

	go func() {
		for {
			in, err := conn.Recv()
			if err != nil {
				if errors.Is(err, mmrpcs.ErrCanceled) || errors.Is(err, io.EOF) {
					log.Println("Recv failed: ", err.Error())
				} else {
					log.Println("Recv failed: ", err.Error())
				}
				break
			}
			log.Println("recv response: ", in.Data()[0])
		}
	}()

	w := packet.Writer()
	w.Write([]byte{1})
	_ = conn.Write(w)

	time.Sleep(time.Second * 10)
}

func Benchmark_Send_Test(b *testing.B) {
	host := "127.0.0.1:6800"
	Serve(b, host)
	conn := DialWithOps(host, &DialOption{
		RemoteNodeId:  "node_0",
		CurrentNodeId: "node_1",
	})
	defer func() {
		_ = conn.Close()
	}()

	go func() {
		for {
			in, err := conn.Recv()
			if err != nil {
				if errors.Is(err, mmrpcs.ErrCanceled) || errors.Is(err, io.EOF) {
					log.Println("Recv failed: ", err.Error())
				} else {
					log.Println("Recv failed: ", err.Error())
				}
				break
			}
			log.Println("recv response: ", in.Data()[0])
		}
	}()

	w := packet.Writer()
	w.Write([]byte{1})
	b.Run("BenchmarkStreamSend", func(b *testing.B) {
		b.ResetTimer()
		b.StartTimer()
		defer b.StopTimer()
		for i := 0; i < b.N; i++ {
			_ = conn.Write(w)
		}
	})

	b.StopTimer()
	time.Sleep(time.Second * 5)
	//_ = conn.Close()
}

func Serve(t TestingT, host string) {
	listener, err := net.Listen("tcp", host)
	assert.NoError(t, err)
	log.Println("start serve...")
	server := new(Server)
	server.Serve(listener, func(conn IServerConn) error {
		for {
			in, err := conn.Recv()
			if err != nil {
				if mmrpcs.IsClosedConnError(err) {
					break
				}
				log.Println("conn read stream failed: ", err.Error())
				break
			}
			log.Println("receive message from client: ", in.Data()[0])
			if err = conn.Send(in); err != nil {
				log.Println("server response failed: ", err.Error())
			}
			in.Return()
		}
		return nil
	})
}

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
}
