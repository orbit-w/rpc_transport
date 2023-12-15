package rpc

import (
	"context"
	"errors"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/golib/modules/unbounded"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"github.com/orbit-w/orbit-net/core/stream_transport/metadata"
	"io"
	"log"
	"runtime/debug"
	"sync/atomic"
	"time"
)

/*
   @Author: orbit-w
   @File: rpc_client
   @2023 12月 周日 17:12
*/

// IClient defines the functions clients need to perform unary and streaming RPCs
// Support two modes
//
//		1: Call
//	 	2: Asynchronous Call
//		3: Shoot
//
// When the transport connection is disconnected, all Call or asynchronous Call requests in the waiting queue
// will return and receive the error 'rpc err: disconnect'
type IClient interface {
	//Shoot is a one-way communication, the sender does not pay attention to the receiver's reply
	Shoot(pid int64, out []byte) error

	// Call performs a unary RPC and returns after the response is received
	// into replyMsg.
	// Support users to use context to cancel blocking status or perform timeout operations
	Call(ctx context.Context, pid int64, out []byte) ([]byte, error)

	// AsyncCall Asynchronous Call requires setting up an asynchronous callback in advance.
	// The callback is handled by a separate goroutine.
	// The caller needs to consider thread safety issues.

	// AsyncCall the safe way to handle asynchronous callbacks is to package the context, replyMsg, and err
	// into a message task and send it to the working goroutine to process the message linearly
	AsyncCall(pid int64, out []byte, ctx any) error

	// AsyncCallC the functionality and precautions of AsyncCallC are similar to AsyncCall.
	// The difference is that AsyncCallC no need to set a global asynchronous callback.
	AsyncCallC(pid int64, out []byte, ctx any, cb func(ctx any, in []byte, err error) error) error

	// Close will close all transport links, causes all subsequent requests to fail,
	// All Call or asynchronous Call requests in the waiting queue will return and receive the error ErrDisconnect 'rpc err: disconnect'
	// Close can be called repeatedly
	Close()
}

type Client struct {
	id         string
	remoteAddr string
	remoteId   string
	state      atomic.Uint32
	seq        atomic.Uint32
	timeout    time.Duration
	stream     stream_transport.IStreamClient
	conn       stream_transport.IClientConn
	codec      Codec
	pending    *Pending
	ch         unbounded.IUnbounded[any]
}

func NewClient(id, remoteId, remoteAddr string) (IClient, error) {
	cli := &Client{
		id:         id,
		remoteAddr: remoteAddr,
		remoteId:   remoteId,
		timeout:    RpcTimeout,
		pending:    new(Pending),
		ch:         unbounded.New[any](2048),
	}

	cli.conn = stream_transport.DialWithOps(remoteAddr, id)
	stream, err := cli.conn.NewStream(metadata.NewMetaContext(context.Background(), map[string]string{
		"nodeId": id,
	}))
	if err != nil {
		_ = cli.conn.Close()
		return nil, err
	}
	cli.stream = stream
	cli.pending.Init(cli, cli.timeout)
	if !cli.state.CompareAndSwap(TypeNone, TypeRunning) {
		_ = cli.stream.CloseSend()
		_ = cli.conn.Close()
		return nil, err
	}
	go cli.loopInput()
	go cli.reader()
	return cli, nil
}

func (c *Client) Close() {
	if c.state.CompareAndSwap(TypeRunning, TypeStopped) {
		if c.stream != nil {
			_ = c.stream.CloseSend()
		}
		time.Sleep(time.Second)
		if c.conn != nil {
			_ = c.conn.Close()
		}
	}
}

func (c *Client) Shoot(pid int64, out []byte) error {
	if c.state.Load() == TypeStopped {
		return mmrpcs.ErrDisconnect
	}
	pack := c.codec.encode(pid, 0, RpcRaw, out)
	return c.stream.Send(pack)
}

func (c *Client) reader() {
	var (
		in  packet.IPacket
		err error
	)

	defer func() {
		if err != nil {
			switch {
			case mmrpcs.IsCancelError(err):
			case errors.Is(err, io.EOF):
			default:
				log.Println("read failed: ", err.Error())
			}
		}

		if c.state.CompareAndSwap(TypeRunning, TypeStopped) {
			_ = c.stream.CloseSend()
			_ = c.conn.Close()
		}

		c.state.CompareAndSwap(TypeNone, TypeStopped)
		if c.ch != nil {
			c.ch.Close()
		}
	}()

	for {
		in, err = c.stream.Recv()
		if err != nil {
			return
		}

		decoder := NewDecoder()
		_ = decoder.Decode(in)

		if err = c.ch.Send(decoder); err != nil {
			if !mmrpcs.IsCancelError(err) {
				log.Println("[Client] [reader] [zq.Write] send in failed")
			}
		}
	}
}

// no blocking
func (c *Client) input(v any) error {
	return c.ch.Send(v)
}

func (c *Client) loopInput() {
	defer func() {
		//TODO: 有没有DeadLock 风险？
		c.pending.RangeAll(func(id uint32) {
			call, ok := c.pending.Pop(id)
			if ok {
				switch {
				case call.IsAsyncInvoker():
					_ = call.Invoke([]byte{}, mmrpcs.ErrDisconnect)
					call.Return()
				default:
					call.Reply([]byte{}, mmrpcs.ErrDisconnect)
				}
			}
		})
		c.pending.OnClose()
		log.Println("[Client] disconnect...")
	}()

	c.ch.Receive(func(msg any) bool {
		c.handleMessage(msg)
		return false
	})
}

func (c *Client) handleMessage(in any) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			log.Println("stack: ", string(debug.Stack()))
		}
	}()

	switch reply := in.(type) {
	case *Decoder:
		call, ok := c.pending.Pop(reply.seq)
		if ok {
			switch {
			case call.IsAsyncInvoker():
				_ = call.Invoke(reply.buf, nil)
				call.Return()
			default:
				call.Reply(reply.buf, nil)
			}
		}
		reply.Return()
	case timeoutListMsg:
		for i := range reply.ids {
			id := reply.ids[i]
			req, ok := c.pending.Pop(id)
			if ok && req.IsAsyncInvoker() {
				_ = req.Invoke([]byte{}, mmrpcs.ErrTimeout)
				req.Return()
			}
		}
	}
}
