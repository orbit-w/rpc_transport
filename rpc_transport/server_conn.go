package rpc_transport

import (
	"context"
	"errors"
	"github.com/orbit-w/meteor/bases/misc/utils"
	"github.com/orbit-w/meteor/modules/mlog"
	"github.com/orbit-w/meteor/modules/net/packet"
	"github.com/orbit-w/meteor/modules/net/transport"
	"go.uber.org/zap"
	"io"
)

type ISession interface {
	Send(seq uint32, category int8, out []byte) error
}

type Conn struct {
	Codec
	conn transport.IConn
	log  *mlog.ZapLogger
}

func NewConn(conn transport.IConn) {
	sConn := newConn(conn)
	sConn.reader()
}

func newConn(conn transport.IConn) *Conn {
	return &Conn{
		conn: conn,
		log:  mlog.NewLogger("[RpcTransport] conn: "),
	}
}

func (c *Conn) Send(seq uint32, category int8, out []byte) error {
	pack := c.Codec.Encode(seq, category, out)
	defer packet.Return(pack)
	return c.conn.Send(pack.Data())
}

func (c *Conn) Close() {
	_ = c.conn.Close()
}

func (c *Conn) reader() {
	defer func() {
		_ = c.conn.Close()
	}()

	ctx := context.Background()

	for {
		in, err := c.conn.Recv(ctx)
		if err != nil {
			switch {
			case transport.IsCancelError(err):
			case errors.Is(err, io.EOF):
			default:
				c.log.Error("conn read failed", zap.Error(err))
			}
			return
		}

		c.handleRequest(in)
	}
}

func (c *Conn) handleRequest(in []byte) {
	req, err := NewRequest(c, in)
	if err != nil {
		c.log.Error("[reader] new request failed", zap.Error(err))
		return
	}

	defer utils.RecoverPanic()

	switch req.Category() {
	case RpcRaw:
		req.IgnoreRsp()
	}

	_ = gRequestHandle(req)
}
