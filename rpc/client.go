package rpc

import (
	"context"
	"errors"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"github.com/orbit-w/orbit-net/core/stream_transport/metadata"
	"github.com/orbit-w/orbit-net/core/unbounded"
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
//
// When the transport connection is disconnected, all Call or asynchronous Call requests in the waiting queue
// will return and receive the error 'rpc err: disconnect'
type IClient interface {
	// Call performs a unary RPC and returns after the response is received
	// into reply.
	// Support users to use context to cancel blocking status or perform timeout operations
	Call(ctx context.Context, pid int64, out []byte) ([]byte, error)

	// AsyncCall Asynchronous Call requires setting up an asynchronous callback in advance.
	// The callback is handled by a separate goroutine.
	// The caller needs to consider thread safety issues.

	// AsyncCall the safe way to handle asynchronous callbacks is to package the context, reply, and err
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

const (
	TypeNone = iota
	TypeRunning
	TypeStopped
)

type Client struct {
	id         string
	remoteAddr string
	remoteId   string
	state      atomic.Uint32
	seq        atomic.Uint32
	timeout    time.Duration
	stream     stream_transport.IStreamClient
	conn       stream_transport.IClientConn
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
		ch:         unbounded.New[any](1024),
	}

	cli.conn = stream_transport.DialWithOps(remoteAddr, id)
	//携带私密信息用于验证
	stream, err := cli.conn.NewStream(metadata.NewMetaContext(context.Background(), map[string]string{
		"nodeId": id,
	}))
	if err != nil {
		_ = cli.conn.Close()
		return nil, err
	}
	cli.stream = stream
	cli.pending.Init(cli.timeout)
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
		if c.conn != nil {
			_ = c.conn.Close()
		}
	}
}

func (c *Client) reader() {
	var (
		in  packet.IPacket
		err error
	)

	defer func() {
		if !IsCancelError(err) {
			log.Println("read failed: ", err.Error())
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

		msg, _ := c.decode(in)
		if err = c.ch.Send(msg); !errors.Is(err, unbounded.ErrCancel) {
			log.Println("reader send in failed: ", err.Error())
		}
	}
}

// no blocking
func (c *Client) input(v any) error {
	return c.ch.Send(v)
}

func (c *Client) loopInput() {
	defer func() {
		c.pending.RangeAll(func(id uint32) {
			req, ok := c.pending.Pop(id)
			if ok {
				switch {
				case req.IsAsyncInvoker():
					_ = req.Invoke([]byte{}, ErrDisconnect)
					req.Return()
				default:
					req.Response([]byte{}, ErrDisconnect)
				}
			}
		})
		c.pending.OnClose()
		log.Println("client disconnect...")
	}()

	c.ch.Receive(func(in any) bool {
		c.handleMessage(in)
		return false
	})
}

func (c *Client) handleMessage(in any) {
	defer func() {
		if v := recover(); v != nil {
			debug.PrintStack()
		}
	}()

	switch msg := in.(type) {
	case message:
		req, ok := c.pending.Pop(msg.seq)
		if ok {
			switch {
			case req.IsAsyncInvoker():
				_ = req.Invoke(msg.reply, nil)
				req.Return()
			default:
				req.Response(msg.reply, nil)
			}
		}
	case timeoutListMsg:
		for i := range msg.ids {
			id := msg.ids[i]
			req, ok := c.pending.Pop(id)
			if ok && req.IsAsyncInvoker() {
				_ = req.Invoke([]byte{}, ErrTimeout)
				req.Return()
			}
		}
	}
}
