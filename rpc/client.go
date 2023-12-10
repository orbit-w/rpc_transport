package rpc

import (
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"log"
	"sync/atomic"
	"time"
)

/*
   @Author: orbit-w
   @File: rpc_client
   @2023 12月 周日 17:12
*/

type Client struct {
	state   atomic.Uint32
	seq     atomic.Uint32
	timeout time.Duration
	stream  stream_transport.IStreamClient
	conn    stream_transport.IClientConn
	pending *Pending
}

func (c *Client) Call(out []byte) ([]byte, error) {
	seq := c.seq.Add(1)
	req := NewRequest(seq)

	c.pending.Push(req)
	if err := c.send(out); err != nil {
		c.pending.Pop(seq)
		req.Return()
		return nil, err
	}

	t := acquireTimer(c.timeout)
	defer releaseTimer(t)

	select {
	case rsp := <-req.Done():
		return rsp.In()
	case <-t.C:
		if _, exist := c.pending.Pop(seq); exist {
			req.Return()
		}
		return nil, ErrTimeout
	}
}

func (c *Client) Close() {
	if c.stream != nil {
		_ = c.stream.CloseSend()
	}

	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) send(out []byte) error {
	pack := packet.Reader(out)
	return c.stream.Send(pack)
}

func (c *Client) reader() {
	var (
		in  packet.IPacket
		err error
	)

	defer func() {
		if !IsCancelError(err) {
			log.Fatalln("read failed: ", err.Error())
		}
	}()

	for {
		in, err = c.stream.Recv()
		if err != nil {
			return
		}

		msg, _ := c.decode(in)
		c.handleMessage(&msg)
	}
}

func (c *Client) handleMessage(msg *message) {
	req, ok := c.pending.Pop(msg.seq)
	if ok {
		switch {
		case req.IsInvoker():
			_ = req.Invoke(msg.reply, nil)
		default:
			req.Response(msg.reply, nil)
		}
	}
}
