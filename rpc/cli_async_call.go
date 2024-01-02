package rpc

var (
	invokeCB       func(ctx any, in []byte, err error) error
	defineInvokeCB = func(ctx any, in []byte, err error) error {
		return nil
	}
)

// SetInvokeCB Set global asynchronous callback
func SetInvokeCB(cb func(ctx any, in []byte, err error) error) {
	invokeCB = cb
}

func (c *Client) AsyncCall(out []byte, ctx any) error {
	if c.state.Load() == TypeStopped {
		return ErrDisconnect
	}
	seq := c.seq.Add(1)
	req := NewCallWithInvoker(seq, NewAsyncInvoker(ctx, nil))
	c.pending.Push(req)
	pack := c.codec.encode(seq, RpcAsyncCall, out)
	defer pack.Return()
	if err := c.conn.Write(pack); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return err
	}
	return nil
}

func (c *Client) AsyncCallC(out []byte, ctx any, cb func(ctx any, in []byte, err error) error) error {
	if c.state.Load() == TypeStopped {
		return ErrDisconnect
	}
	seq := c.seq.Add(1)
	req := NewCallWithInvoker(seq, NewAsyncInvoker(ctx, cb))
	c.pending.Push(req)
	pack := c.codec.encode(seq, RpcAsyncCall, out)
	defer pack.Return()
	if err := c.conn.Write(pack); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return err
	}
	return nil
}
