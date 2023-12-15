package rpc

import (
	"github.com/orbit-w/mmrpc/rpc/callb"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	"log"
)

var (
	invokeCB func(ctx any, in []byte, err error) error
)

// SetInvokeCB Set global asynchronous callback
func SetInvokeCB(cb func(ctx any, in []byte, err error) error) {
	invokeCB = cb
}

func (c *Client) AsyncCall(pid int64, out []byte, ctx any) error {
	if c.state.Load() == TypeStopped {
		return mmrpcs.ErrDisconnect
	}
	seq := c.seq.Add(1)
	req := callb.NewCallWithInvoker(seq, NewAsyncInvoker(ctx, nil))
	log.Println("AsyncCall start ...")
	c.pending.Push(req)
	log.Println("AsyncCall push ...")
	pack := c.codec.encode(pid, seq, RpcAsyncCall, out)
	defer pack.Return()
	if err := c.stream.Send(pack); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return err
	}
	log.Println("AsyncCall...")
	return nil
}

func (c *Client) AsyncCallC(pid int64, out []byte, ctx any, cb func(ctx any, in []byte, err error) error) error {
	if c.state.Load() == TypeStopped {
		return mmrpcs.ErrDisconnect
	}
	seq := c.seq.Add(1)
	req := callb.NewCallWithInvoker(seq, NewAsyncInvoker(ctx, cb))
	c.pending.Push(req)
	pack := c.codec.encode(pid, seq, RpcAsyncCall, out)
	defer pack.Return()
	if err := c.stream.Send(pack); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return err
	}
	return nil
}
