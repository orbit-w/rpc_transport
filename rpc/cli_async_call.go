package rpc

import (
	"github.com/orbit-w/mmrpc/rpc/callb"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
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
	c.pending.Push(req)
	pack := c.codec.encode(pid, seq, RpcAsyncCall, out)
	if err := c.stream.Send(pack); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return err
	}
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
	if err := c.stream.Send(pack); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return err
	}
	return nil
}
