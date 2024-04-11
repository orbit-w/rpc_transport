package rpc_transport

import (
	"errors"
	"github.com/orbit-w/golib/modules/net/transport"
	"io"
	"runtime/debug"
)

type ISession interface {
	Send(seq uint32, category int8, out []byte) error
}

type Conn struct {
	Codec
	conn transport.IConn
}

func NewConn(conn transport.IConn) {
	sConn := Conn{}
	sConn.conn = conn
	sConn.reader()
}

func (c *Conn) Send(seq uint32, category int8, out []byte) error {
	pack := c.Codec.encode(seq, category, out)
	defer pack.Return()
	return c.conn.Send(pack.Data())
}

func (c *Conn) Close() {
	_ = c.conn.Close()
}

func (c *Conn) reader() {
	defer func() {
		_ = c.conn.Close()
	}()
	for {
		in, err := c.conn.Recv()
		if err != nil {
			switch {
			case transport.IsCancelError(err):
			case errors.Is(err, io.EOF):
			default:
				SugarLogger().Error("conn read failed: ", err.Error())
			}
			return
		}

		c.handleRequest(in)
	}
}

func (c *Conn) handleRequest(in []byte) {
	req, err := NewRequest(c, in)
	if err != nil {
		SugarLogger().Error("[ServerConn] [reader] new request failed: ", err.Error())
		return
	}
	defer func() {
		if r := recover(); r != nil {
			SugarLogger().Error(r)
			SugarLogger().Error("stack: ", string(debug.Stack()))
		}
	}()

	switch req.Category() {
	case RpcRaw:
		req.IgnoreRsp()
	}

	_ = gRequestHandle(req)
}
