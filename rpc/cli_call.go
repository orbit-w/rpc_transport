package rpc

import (
	"context"
	"errors"
	"github.com/orbit-w/golib/modules/transport"
	"github.com/orbit-w/mmrpc/rpc/callb"
)

func (c *Client) Call(ctx context.Context, pid int64, out []byte) ([]byte, error) {
	if c.state.Load() == TypeStopped {
		return nil, ErrDisconnect
	}
	seq := c.seq.Add(1)
	call := callb.NewCall(seq)

	c.pending.Push(call)
	pack := c.codec.encode(pid, seq, RpcCall, out)
	defer pack.Return()
	if err := c.conn.Write(pack); err != nil {
		c.pending.Pop(seq)
		call.Return()
		return nil, err
	}
	select {
	case reply := <-call.Done():
		data, err := reply.In()
		call.Return()
		reply.Return()
		if err != nil {
			switch {
			case transport.IsCancelError(err):
				return data, err
			default:
				return data, NewRpcError(err)
			}
		}
		return data, err
	case <-ctx.Done():
		if _, exist := c.pending.Pop(seq); exist {
			call.Return()
		}
		err := ctx.Err()
		switch {
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
			return nil, err
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

type RingBuffer[v any] struct {
}
