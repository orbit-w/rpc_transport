package rpc_transport

import (
	"context"
	"errors"
	"github.com/orbit-w/meteor/modules/net/packet"
	"github.com/orbit-w/meteor/modules/net/transport"
)

func (c *Client) Call(ctx context.Context, out []byte) ([]byte, error) {
	if c.state.Load() == TypeStopped {
		return nil, ErrDisconnect
	}
	seq := c.seq.Add(1)
	call := NewCall(seq)

	c.pending.Push(call)
	pack := c.codec.Encode(seq, RpcCall, out)
	defer packet.Return(pack)
	if err := c.conn.Send(pack.Data()); err != nil {
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
