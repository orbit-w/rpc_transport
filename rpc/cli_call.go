package rpc

import (
	"context"
	"errors"
	"github.com/orbit-w/mmrpc/rpc/callb"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
)

func (c *Client) Call(ctx context.Context, pid int64, out []byte) ([]byte, error) {
	if c.state.Load() == TypeStopped {
		return nil, mmrpcs.ErrDisconnect
	}
	seq := c.seq.Add(1)
	call := callb.NewCall(seq)

	c.pending.Push(call)
	pack := c.codec.encode(pid, seq, RpcCall, out)
	if err := c.stream.Send(pack); err != nil {
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
			case mmrpcs.IsCancelError(err):
				return data, mmrpcs.ErrCanceled
			default:
				return data, mmrpcs.NewRpcError(err)
			}
		}
		return data, err
	case <-ctx.Done():
		if _, exist := c.pending.Pop(seq); exist {
			call.Return()
		}
		err := ctx.Err()
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return nil, mmrpcs.ErrDeadlineExceeded
		case errors.Is(err, context.Canceled):
			return nil, mmrpcs.ErrCanceled
		default:
			return nil, mmrpcs.NewRpcError(err)
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
