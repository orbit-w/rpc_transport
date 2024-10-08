package rpc_transport

import (
	"context"
	"errors"
	"fmt"
	"github.com/orbit-w/meteor/modules/mlog"
	"github.com/orbit-w/meteor/modules/net/packet"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/meteor/modules/unbounded"
	"go.uber.org/zap"
	"io"
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
// will return and receive the error 'rpc_transport err: disconnect'
type IClient interface {
	//Shoot is a one-way communication, the sender does not pay attention to the receiver's reply
	Shoot(out []byte) error

	// Call performs a unary RPC and returns after the response is received
	// into replyMsg.
	// Support users to use context to cancel blocking status or perform timeout operations
	Call(ctx context.Context, out []byte) ([]byte, error)

	// AsyncCall Asynchronous Call requires setting up an asynchronous callback in advance.
	// The callback is handled by a separate goroutine.
	// The caller needs to consider thread safety issues.

	// AsyncCall the safe way to handle asynchronous callbacks is to package the context, replyMsg, and err
	// into a message task and send it to the working goroutine to process the message linearly
	AsyncCall(out []byte, ctx any) error

	// AsyncCallC the functionality and precautions of AsyncCallC are similar to AsyncCall.
	// The difference is that AsyncCallC no need to set a global asynchronous callback.
	AsyncCallC(out []byte, ctx any, cb func(ctx any, in []byte, err error) error) error

	// Close will close all transport links, causes all subsequent requests to fail,
	// All Call or asynchronous Call requests in the waiting queue will return and receive the error ErrDisconnect 'rpc_transport err: disconnect'
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
	conn       transport.IConn
	codec      Codec
	pending    *Pending
	ch         unbounded.IUnbounded[any]
	log        *mlog.ZapLogger
}

type DialOption struct {
	DisconnectHandler func()
}

func Dial(id, remoteId, addr string, ops ...*DialOption) (IClient, error) {
	cli := &Client{
		id:         id,
		remoteAddr: addr,
		remoteId:   remoteId,
		timeout:    RpcTimeout,
		pending:    new(Pending),
		ch:         unbounded.New[any](2048),
		log:        mlog.NewLogger("[RpcTransport] client: "),
	}

	cli.conn = transport.DialWithOps(cli.remoteAddr, cli.parseOpToTransportOp(ops...))
	cli.pending.Init(cli, cli.timeout)
	if !cli.state.CompareAndSwap(TypeNone, TypeRunning) {
		_ = cli.conn.Close()
		return nil, ErrDisconnect
	}
	go cli.loopInput()
	go cli.reader()
	return cli, nil
}

func (c *Client) Close() {
	if c.state.CompareAndSwap(TypeRunning, TypeStopped) {
		if c.conn != nil {
			_ = c.conn.Close()
		}
	}
}

func (c *Client) Shoot(out []byte) error {
	if c.state.Load() == TypeStopped {
		return ErrDisconnect
	}
	pack := c.codec.Encode(0, RpcRaw, out)
	defer packet.Return(pack)
	return c.conn.Send(pack.Data())
}

func (c *Client) reader() {
	var (
		in  []byte
		err error
		ctx = context.Background()
	)

	defer func() {
		if err != nil {
			switch {
			case transport.IsCancelError(err):
			case errors.Is(err, io.EOF):
			default:
				c.log.Error("read failed", zap.Error(err))
			}
		}

		if c.state.CompareAndSwap(TypeRunning, TypeStopped) {
			_ = c.conn.Close()
		}

		c.state.CompareAndSwap(TypeNone, TypeStopped)
		if c.ch != nil {
			c.ch.Close()
		}
	}()

	for {
		in, err = c.conn.Recv(ctx)
		if err != nil {
			return
		}
		decoder := NewDecoder()
		_ = decoder.Decode(in)

		if err = c.ch.Send(decoder); err != nil {
			if !transport.IsCancelError(err) {
				c.log.Error("[zq.Write] send in failed", zap.Error(err))
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
		c.pending.RangeAll(func(id uint32) {
			call, ok := c.pending.Pop(id)
			if ok {
				switch {
				case call.IsAsyncInvoker():
					_ = call.Invoke([]byte{}, ErrDisconnect)
					call.Return()
				default:
					call.Reply([]byte{}, ErrDisconnect)
				}
			}
		})
		c.pending.OnClose()
		c.log.Info("disconnect...")
	}()

	c.ch.Receive(func(msg any) bool {
		c.handleMessage(msg)
		return false
	})
}

func (c *Client) handleMessage(in any) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			fmt.Println("Stack trace:")
			debug.PrintStack()
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
				_ = req.Invoke([]byte{}, ErrTimeout)
				req.Return()
			}
		}
	}
}

func (c *Client) parseOpToTransportOp(ops ...*DialOption) *transport.DialOption {
	var dh func()
	if len(ops) > 0 {
		op := ops[0]
		dh = op.DisconnectHandler
	}
	return &transport.DialOption{
		DisconnectHandler: dh,
	}
}
