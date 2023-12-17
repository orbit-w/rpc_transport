package rpc

import (
	"errors"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	"github.com/orbit-w/mmrpc/rpc/transport"
	"io"
	"log"
	"runtime/debug"
)

type ISession interface {
	Send(pid int64, seq uint32, category int8, out []byte) error
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

func (c *Conn) Send(pid int64, seq uint32, category int8, out []byte) error {
	pack := c.Codec.encode(pid, seq, category, out)
	defer pack.Return()
	return c.conn.Write(pack)
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
			case mmrpcs.IsCancelError(err):
			case errors.Is(err, io.EOF):
			default:
				log.Println("conn read failed: ", err.Error())
			}
			return
		}

		c.handleRequest(in)
	}
}

func (c *Conn) handleRequest(in packet.IPacket) {
	req, err := NewRequest(c, in)
	if err != nil {
		log.Println("[ServerConn] [reader] new request failed: ", err.Error())
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			log.Println("stack: ", string(debug.Stack()))
		}
	}()

	switch req.Category() {
	case RpcRaw:
		req.IgnoreRsp()
	}

	_ = gRequestHandle(req)
}
