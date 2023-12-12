package rpc

import (
	"context"
	"errors"
)

func (c *Client) Call(ctx context.Context, pid int64, out []byte) ([]byte, error) {
	if c.state.Load() == TypeStopped {
		return nil, ErrDisconnect
	}
	seq := c.seq.Add(1)
	req := NewRequest(seq)

	c.pending.Push(req)
	pack := c.encode(pid, seq, RpcCall, out)
	if err := c.stream.Send(pack); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return nil, err
	}

	select {
	case rsp := <-req.Done():
		reply, err := rsp.In()
		req.Return()
		rsp.Return()
		switch {
		case IsCancelError(err):
			return reply, ErrCanceled
		default:
			return reply, NewRpcError(err)
		}
	case <-ctx.Done():
		if _, exist := c.pending.Pop(seq); exist {
			req.Return()
		}
		err := ctx.Err()
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return nil, ErrDeadlineExceeded
		case errors.Is(err, context.Canceled):
			return nil, ErrCanceled
		default:
			return nil, NewRpcError(err)
		}
	}
}

func wrapCtx(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	ctx, _ = context.WithTimeout(context.Background(), RpcTimeout)
	return ctx
}
